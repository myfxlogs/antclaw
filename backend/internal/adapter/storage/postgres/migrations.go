// Package postgres: 嵌入式 SQL 迁移加载器。
//
// 设计取舍：
//   - 不引入 goose / atlas 等外部工具：单一二进制启动即可建表
//   - 用 schema_migrations(filename) 跟踪已应用迁移，文件名升序串行执行
//   - 每个 .sql 文件作为一次 pool.Exec 提交；要求文件本身是幂等的
//     （CREATE TABLE IF NOT EXISTS / create_hypertable(if_not_exists=>TRUE) /
//      DROP TRIGGER IF EXISTS ... + CREATE TRIGGER 等）
//   - 仅当 .sql 文件确为新增（schema_migrations 中没有）时才执行；
//     已应用 → 跳过；执行失败 → 立即返回错误（不静默吞掉），
//     便于发现真实的 schema 漂移。
package postgres

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed all:migrations
var migrationFS embed.FS

// splitStatements 将一个 .sql 文件文本按顶层分号切成独立语句，
// 正确处理：单引号字符串 / 双引号标识符 / dollar-quoted 块 / 行注释 / 块注释。
// 之所以需要逐条执行：TimescaleDB 的 CREATE MATERIALIZED VIEW WITH (timescaledb.continuous)
// 不允许出现在事务块里，而 pgx 的多语句 simple query 会被当作隐式事务。
func splitStatements(src string) []string {
	var out []string
	var buf strings.Builder
	type mode int
	const (
		mNormal mode = iota
		mLine
		mBlock
		mSingle
		mDouble
		mDollar
	)
	st := mNormal
	dollarTag := ""
	r := []rune(src)
	for i := 0; i < len(r); i++ {
		c := r[i]
		next := func() rune {
			if i+1 < len(r) {
				return r[i+1]
			}
			return 0
		}
		switch st {
		case mLine:
			buf.WriteRune(c)
			if c == '\n' {
				st = mNormal
			}
		case mBlock:
			buf.WriteRune(c)
			if c == '*' && next() == '/' {
				buf.WriteRune('/')
				i++
				st = mNormal
			}
		case mSingle:
			buf.WriteRune(c)
			if c == '\'' {
				st = mNormal
			}
		case mDouble:
			buf.WriteRune(c)
			if c == '"' {
				st = mNormal
			}
		case mDollar:
			buf.WriteRune(c)
			if c == '$' {
				// 看是否匹配关闭 tag
				end := i + len(dollarTag)
				if end <= len(r) && string(r[i:end+1]) == dollarTag {
					for j := i + 1; j <= end; j++ {
						buf.WriteRune(r[j])
					}
					i = end
					st = mNormal
				}
			}
		case mNormal:
			if c == '-' && next() == '-' {
				buf.WriteRune(c)
				st = mLine
				continue
			}
			if c == '/' && next() == '*' {
				buf.WriteRune(c)
				buf.WriteRune('*')
				i++
				st = mBlock
				continue
			}
			if c == '\'' {
				buf.WriteRune(c)
				st = mSingle
				continue
			}
			if c == '"' {
				buf.WriteRune(c)
				st = mDouble
				continue
			}
			if c == '$' {
				// 可能是 dollar-quote 起始 $tag$
				j := i + 1
				for j < len(r) && (r[j] == '_' || (r[j] >= 'a' && r[j] <= 'z') || (r[j] >= 'A' && r[j] <= 'Z') || (r[j] >= '0' && r[j] <= '9')) {
					j++
				}
				if j < len(r) && r[j] == '$' {
					dollarTag = "$" + string(r[i+1:j]) + "$"
					for k := i; k <= j; k++ {
						buf.WriteRune(r[k])
					}
					i = j
					st = mDollar
					continue
				}
				buf.WriteRune(c)
				continue
			}
			if c == ';' {
				stmt := strings.TrimSpace(buf.String())
				if stmt != "" {
					out = append(out, stmt)
				}
				buf.Reset()
				continue
			}
			buf.WriteRune(c)
		}
	}
	if last := strings.TrimSpace(buf.String()); last != "" {
		out = append(out, last)
	}
	return out
}

// stripGooseDown 截断 "-- +goose Down" 及其后的所有内容，
// 只保留 Up 段；大小写不敏感，行首允许任意空白。
func stripGooseDown(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		t := strings.ToLower(strings.TrimSpace(ln))
		if strings.HasPrefix(t, "-- +goose down") {
			return strings.Join(lines[:i], "\n")
		}
	}
	return s
}

// RunMigrations 在 API 启动时执行 db/migrations/*.sql。
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations dir: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	rows, err := pool.Query(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		applied[name] = true
	}
	rows.Close()

	for _, name := range files {
		if applied[name] {
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read embedded migration %s: %w", name, err)
		}
		// 迁移文件原本由 goose 管理，只跑 "-- +goose Up" 段；
		// 没有 goose 运行器时，必须显式截断，避免把 Down 段（含 DROP）当真执行。
		sql := stripGooseDown(string(body))

		for _, stmt := range splitStatements(sql) {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("apply migration %s: %w\n-- statement:\n%s", name, err, stmt)
			}
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations(filename) VALUES ($1) ON CONFLICT DO NOTHING`,
			name); err != nil {
			return fmt.Errorf("record applied migration %s: %w", name, err)
		}
	}
	return nil
}

