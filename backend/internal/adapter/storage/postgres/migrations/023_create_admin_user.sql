-- +goose Up
-- Create default admin user
-- Email: a@1.com
-- Password: 12345678 (hashed with Argon2ID)

INSERT INTO users (
    id,
    email,
    email_verified_at,
    display_name,
    password_hash,
    password_version,
    role,
    status,
    locale,
    timezone
) VALUES (
    gen_random_uuid(),
    'a@1.com',
    NOW(),
    'Admin',
    '$argon2id$v=19$m=65536,t=3,p=2$kEJKtjOJFUwXND7TabdPAA$Ae4aApT7UeFFFrrJIYE+QmfBN24NsuE0fcBbnizrExk',
    1,
    'admin',
    'active',
    'zh-CN',
    'Asia/Shanghai'
) ON CONFLICT (email) DO NOTHING;

-- +goose Down
-- Remove default admin user
DELETE FROM users WHERE email = 'a@1.com';
