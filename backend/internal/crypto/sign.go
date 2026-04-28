package crypto

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

// 接口签名：HMAC-SHA256(session_key, timestamp || "\n" || nonce || "\n" || body)
//   - session_key 由前端从 JWT token 派生：sha256(token)；服务端可同样派生
//   - timestamp：unix seconds，允许 ±5 分钟漂移
//   - nonce：客户端生成的唯一字符串；服务端用 Redis 短期存储防重放
//
// 该签名独立于 hybrid 加密，提供"请求来自合法 session 且未被篡改"的额外保证。

const (
	maxClockSkew = 5 * 60          // 5 分钟
	nonceTTL     = 10 * time.Minute // > maxClockSkew 即可
)

// SessionKey 从 JWT token 派生出客户端 / 服务端共享的 HMAC 密钥。
// token 不出网，因此即便派生函数公开，密钥仍是 session 私有的。
func SessionKey(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// VerifyRequestSignature 验证请求签名 + 时间戳 + nonce 防重放。
// rdb 用于 nonce 去重；可为 nil（仅校验签名与时间）。
func VerifyRequestSignature(
	ctx context.Context,
	rdb *redisv9.Client,
	token string,
	tsStr string,
	nonce string,
	body []byte,
	signatureHex string,
) error {
	if token == "" {
		return errors.New("missing session token")
	}
	if tsStr == "" || nonce == "" || signatureHex == "" {
		return errors.New("missing signature headers")
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	now := time.Now().Unix()
	if ts < now-maxClockSkew || ts > now+maxClockSkew {
		return fmt.Errorf("timestamp out of window")
	}

	mac := hmac.New(sha256.New, SessionKey(token))
	mac.Write([]byte(tsStr))
	mac.Write([]byte("\n"))
	mac.Write([]byte(nonce))
	mac.Write([]byte("\n"))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(signatureHex)) {
		return errors.New("signature mismatch")
	}

	if rdb != nil {
		key := "sigreplay:" + nonce
		// SETNX，防 nonce 重放
		ok, err := rdb.SetNX(ctx, key, "1", nonceTTL).Result()
		if err != nil {
			return fmt.Errorf("nonce check: %w", err)
		}
		if !ok {
			return errors.New("nonce already used")
		}
	}
	return nil
}
