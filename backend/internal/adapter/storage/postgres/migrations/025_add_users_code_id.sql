-- +goose Up
-- 用户数字 ID（code_id）：人类可记忆的账号，与 UUID 主键并存。
-- 字符集：首位 {1,2,3,5,6,8,9}，其余位 {0,1,2,3,5,6,8,9}（避开 4 和 7）。
-- 默认 5 位，列上不限位数以便后期扩展。

ALTER TABLE users ADD COLUMN IF NOT EXISTS code_id TEXT;

-- 仅对已分配的 code_id 唯一；NULL 不参与（部分索引）。
CREATE UNIQUE INDEX IF NOT EXISTS users_code_id_uq
    ON users(code_id) WHERE code_id IS NOT NULL;

-- 字符与长度（5-10 位）双重校验；正则用枚举字符类，无歧义。
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_code_id_format;
ALTER TABLE users ADD CONSTRAINT users_code_id_format CHECK (
    code_id IS NULL
    OR code_id ~ '^[1235689][01235689]{4,9}$'
);

-- 回填用辅助函数：基于 [1-3,5-6,8-9] 字符集生成 n 位 code_id。
CREATE OR REPLACE FUNCTION gen_user_code_id(n INT DEFAULT 5) RETURNS TEXT AS $$
DECLARE
    first_chars CONSTANT TEXT := '1235689';
    rest_chars  CONSTANT TEXT := '01235689';
    result TEXT;
    i INT;
BEGIN
    IF n < 5 THEN n := 5; END IF;
    result := substr(first_chars, 1 + floor(random() * length(first_chars))::INT, 1);
    FOR i IN 2..n LOOP
        result := result || substr(rest_chars, 1 + floor(random() * length(rest_chars))::INT, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql VOLATILE;

-- 回填所有现存且 code_id 为空的用户，唯一冲突时重试。
DO $$
DECLARE
    u RECORD;
    cid TEXT;
    attempts INT;
BEGIN
    FOR u IN SELECT id FROM users WHERE code_id IS NULL AND deleted_at IS NULL LOOP
        attempts := 0;
        LOOP
            cid := gen_user_code_id(5);
            BEGIN
                UPDATE users SET code_id = cid WHERE id = u.id;
                EXIT;
            EXCEPTION WHEN unique_violation THEN
                attempts := attempts + 1;
                IF attempts > 20 THEN
                    RAISE EXCEPTION 'cannot allocate code_id for user % after 20 attempts', u.id;
                END IF;
            END;
        END LOOP;
    END LOOP;
END$$;

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_code_id_format;
DROP INDEX IF EXISTS users_code_id_uq;
ALTER TABLE users DROP COLUMN IF EXISTS code_id;
DROP FUNCTION IF EXISTS gen_user_code_id(INT);
