-- 030: devices 表 — 用户设备信息上报
-- 对应 proto: antclaw/v1/device.proto DeviceInfo
CREATE TABLE IF NOT EXISTS devices (
    device_id      VARCHAR(255) PRIMARY KEY,         -- 客户端生成的持久化设备 ID
    model          VARCHAR(128) NOT NULL DEFAULT '',  -- 设备型号
    brand          VARCHAR(64)  NOT NULL DEFAULT '',  -- 品牌
    os_version     VARCHAR(32)  NOT NULL DEFAULT '',  -- 操作系统版本
    os_type        VARCHAR(16)  NOT NULL DEFAULT '',  -- android / ios / web
    app_version    VARCHAR(32)  NOT NULL DEFAULT '',  -- App 版本号
    build_number   VARCHAR(32)  NOT NULL DEFAULT '',  -- 构建号
    screen_width   INTEGER      NOT NULL DEFAULT 0,   -- 屏幕宽度
    screen_height  INTEGER      NOT NULL DEFAULT 0,   -- 屏幕高度
    network_type   VARCHAR(32)  NOT NULL DEFAULT '',  -- wifi / cellular / ethernet
    timezone       VARCHAR(64)  NOT NULL DEFAULT '',  -- 时区
    locale         VARCHAR(16)  NOT NULL DEFAULT '',  -- 语言区域
    manufacturer   VARCHAR(64)  NOT NULL DEFAULT '',  -- 制造商
    fingerprint    VARCHAR(255) NOT NULL DEFAULT '',  -- 设备指纹
    user_id        UUID         REFERENCES users(id) ON DELETE SET NULL,  -- 关联用户
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_devices_os_type ON devices(os_type, updated_at DESC);
