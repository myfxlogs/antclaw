// codeid.go —— 用户数字 ID（code_id）生成与校验。
//
// 规格：
//   - 首位字符：{1,2,3,5,6,8,9}     7 种（不为 0、不含 4/7）
//   - 其余字符：{0,1,2,3,5,6,8,9}    8 种
//   - 默认 5 位，可参数化扩展
//   - 安全随机源 crypto/rand，避免顺序可预测
//
// 容量参考：n=5 → 7×8⁴ = 28,672；n=6 → 229,376；n=7 → 1,835,008。
package auth

import (
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
)

const (
	// CodeIDDefaultDigits 注册时默认生成的位数；后期改大不影响已有 ID。
	CodeIDDefaultDigits = 5
	// CodeIDMinDigits 列约束允许的最小位数。
	CodeIDMinDigits = 5
	// CodeIDMaxDigits 列约束允许的最大位数。
	CodeIDMaxDigits = 10

	codeIDFirstChars = "1235689"
	codeIDRestChars  = "01235689"
)

// codeIDRegex 与 SQL CHECK 约束保持一致。
var codeIDRegex = regexp.MustCompile(`^[1235689][01235689]{4,9}$`)

// ErrInvalidCodeID 校验失败。
var ErrInvalidCodeID = errors.New("invalid code_id: must be 5-10 digits, exclude 4/7, not start with 0")

// GenerateCodeID 用安全随机源生成 n 位 code_id。
// n < CodeIDMinDigits 时按 CodeIDMinDigits 处理，> CodeIDMaxDigits 时按上限处理。
func GenerateCodeID(n int) (string, error) {
	if n < CodeIDMinDigits {
		n = CodeIDMinDigits
	}
	if n > CodeIDMaxDigits {
		n = CodeIDMaxDigits
	}
	buf := make([]byte, n)
	first, err := pickByte(codeIDFirstChars)
	if err != nil {
		return "", err
	}
	buf[0] = first
	for i := 1; i < n; i++ {
		c, err := pickByte(codeIDRestChars)
		if err != nil {
			return "", err
		}
		buf[i] = c
	}
	return string(buf), nil
}

// ValidateCodeID 严格校验：长度、字符集、首位非 0、不含 4/7。
func ValidateCodeID(s string) error {
	if !codeIDRegex.MatchString(s) {
		return ErrInvalidCodeID
	}
	return nil
}

// IsAllDigits 用于登录入口智能识别：判断 identifier 是否全数字。
// 注：仅判定数字组成，不严格按 code_id 字符集（容忍历史/未来差异）。
func IsAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func pickByte(alphabet string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
	if err != nil {
		return 0, err
	}
	return alphabet[n.Int64()], nil
}
