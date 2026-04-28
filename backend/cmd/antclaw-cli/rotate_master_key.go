// rotate_master_key.go implements master key rotation for encrypted secrets.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
)

// RotatePlan holds configuration for key rotation.
type RotatePlan struct {
	OldMaster string
	NewMaster string
	Tables    []string
	DryRun    bool
}

// Report holds rotation results.
type Report struct {
	PerTable                    map[string]TableStat
	OK, Fail, Skip, WouldRotate int
}

// TableStat holds per-table statistics.
type TableStat struct {
	OK, Fail, Skip int
	FailedRows     []string
}

// Run executes the rotation plan.
func Run(ctx context.Context, plan RotatePlan, pool *pgxpool.Pool, logger *slog.Logger) (Report, error) {
	oldBox, err := cryptopkg.NewSecretBox(plan.OldMaster)
	if err != nil {
		return Report{}, fmt.Errorf("invalid old master key: %w", err)
	}
	newBox, err := cryptopkg.NewSecretBox(plan.NewMaster)
	if err != nil {
		return Report{}, fmt.Errorf("invalid new master key: %w", err)
	}

	if plan.OldMaster == plan.NewMaster {
		return Report{}, errors.New("old and new master keys are identical")
	}

	report := Report{
		PerTable: make(map[string]TableStat),
	}

	for _, table := range plan.Tables {
		stat, err := rotateTable(ctx, pool, oldBox, newBox, table, plan.DryRun, logger)
		report.PerTable[table] = stat
		report.OK += stat.OK
		report.Fail += stat.Fail
		report.Skip += stat.Skip
		if plan.DryRun {
			report.WouldRotate += stat.OK // In dry-run, OK means "would rotate"
		}
		if err != nil {
			logger.Warn("table rotation error", "table", table, "error", err)
		}
	}

	return report, nil
}

// tableConfig defines the column names for each table.
func getTableConfig(table string) (idCol string, err error) {
	switch table {
	case "data_source_configs":
		return "source_id", nil
	case "system_ai_configs":
		return "provider_id", nil
	default:
		return "", fmt.Errorf("unknown table: %s", table)
	}
}

func rotateTable(ctx context.Context, pool *pgxpool.Pool, oldBox, newBox *cryptopkg.SecretBox, table string, dryRun bool, logger *slog.Logger) (TableStat, error) {
	idCol, err := getTableConfig(table)
	if err != nil {
		return TableStat{}, err
	}

	stat := TableStat{}

	// Query rows with secrets
	query := fmt.Sprintf(`
		SELECT %s, secret_ciphertext, secret_salt, secret_nonce 
		FROM %s WHERE has_secret = TRUE`, idCol, table)
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return stat, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var ct, salt, nonce []byte
		if err := rows.Scan(&id, &ct, &salt, &nonce); err != nil {
			stat.Fail++
			stat.FailedRows = append(stat.FailedRows, fmt.Sprintf("scan_error:%v", err))
			continue
		}

		// Decrypt with old key
		plaintext, err := oldBox.Open(ct, salt, nonce)
		if err != nil {
			stat.Fail++
			stat.FailedRows = append(stat.FailedRows, id)
			logger.Error("decrypt failed", "table", table, "id", id, "error", err)
			continue
		}

		if dryRun {
			stat.OK++ // In dry-run, count as "would rotate"
			fmt.Printf("[dry-run] would rotate %s in %s\n", id, table)
			continue
		}

		// Re-encrypt with new key
		newCT, newSalt, newNonce, err := newBox.Seal(plaintext)
		if err != nil {
			stat.Fail++
			stat.FailedRows = append(stat.FailedRows, id)
			logger.Error("seal failed", "table", table, "id", id, "error", err)
			continue
		}

		// Update row
		update := fmt.Sprintf(`
			UPDATE %s SET secret_ciphertext=$1, secret_salt=$2, secret_nonce=$3, updated_at=NOW() 
			WHERE %s=$4`, table, idCol)
		_, err = pool.Exec(ctx, update, newCT, newSalt, newNonce, id)
		if err != nil {
			stat.Fail++
			stat.FailedRows = append(stat.FailedRows, id)
			logger.Error("update failed", "table", table, "id", id, "error", err)
			continue
		}

		stat.OK++
		fmt.Printf("OK  %s\n", id)
	}

	if err := rows.Err(); err != nil {
		return stat, fmt.Errorf("rows iteration failed: %w", err)
	}

	return stat, nil
}

// rotateMasterKeyCmd is the entry point for the rotate-master-key subcommand.
func rotateMasterKeyCmd(args []string) int {
	var plan RotatePlan
	var tables string

	// Simple flag parsing
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--old":
			if i+1 < len(args) {
				plan.OldMaster = args[i+1]
				i++
			}
		case "--new":
			if i+1 < len(args) {
				plan.NewMaster = args[i+1]
				i++
			}
		case "--tables":
			if i+1 < len(args) {
				tables = args[i+1]
				i++
			}
		case "--dry-run":
			plan.DryRun = true
		}
	}

	if plan.OldMaster == "" || plan.NewMaster == "" {
		fmt.Fprintln(os.Stderr, "Usage: antclaw-cli rotate-master-key --old=<base64> --new=<base64> [--dry-run] [--tables=t1,t2]")
		return 1
	}

	// Default tables
	if tables == "" {
		plan.Tables = []string{"data_source_configs", "system_ai_configs"}
	} else {
		plan.Tables = splitTables(tables)
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), getDBConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	report, err := Run(context.Background(), plan, pool, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rotation failed: %v\n", err)
		return 1
	}

	// Print summary
	fmt.Println("---")
	if plan.DryRun {
		fmt.Printf("Total: %d would rotate, %d fail, %d skip\n", report.WouldRotate, report.Fail, report.Skip)
	} else {
		fmt.Printf("Total: %d OK, %d FAIL, %d SKIP\n", report.OK, report.Fail, report.Skip)
	}

	if report.Fail > 0 {
		return 1
	}
	return 0
}

func splitTables(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func getDBConnString() string {
	host := os.Getenv("ANTCLAW_DB_HOST")
	if host == "" {
		host = "postgres"
	}
	port := os.Getenv("ANTCLAW_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("ANTCLAW_DB_USER")
	if user == "" {
		user = "antclaw"
	}
	pass := os.Getenv("ANTCLAW_DB_PASSWORD")
	if pass == "" {
		pass = "antclaw"
	}
	dbname := os.Getenv("ANTCLAW_DB_NAME")
	if dbname == "" {
		dbname = "antclaw"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
}
