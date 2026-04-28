package admin

import "errors"

// ErrPasswordPolicy 表示新密码不满足策略（长度等）。
var ErrPasswordPolicy = errors.New("password does not meet policy")
