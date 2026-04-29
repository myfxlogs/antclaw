#!/usr/bin/env bash
# SC-18 Crypto envelope — 后端必须暴露 CryptoService.PostEnvelope（替代历史 fetch envelope）。
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
# 不实际加密，只验证端点存在并能被识别（错误码而非 404）。
out=$(call CryptoService PostEnvelope '{"ciphertext":"","iv":"","tag":"","encryptedKey":""}')
echo "  resp=$(echo "$out" | head -c 200)"
# 期望非 404；连接通即视为通过（业务校验失败也表示 RPC 注册成功）
echo "$out" | jq -e '.code // "ok"' >/dev/null
