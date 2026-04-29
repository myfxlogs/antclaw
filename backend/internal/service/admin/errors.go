package admin

import "errors"

// ErrPasswordPolicy 表示新密码不满足策略（长度等）。
var ErrPasswordPolicy = errors.New("password does not meet policy")

// ErrCodeIDTaken 指定的 code_id 已被其它用户占用。
var ErrCodeIDTaken = errors.New("code_id already in use")
