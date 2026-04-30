#!/usr/bin/env bash
# 对 Connect/RPC 端点做端到端 smoke 测试。
# 仅校验 HTTP 200 响应（数据正确性由各 handler 单测承担）。
# 用法：API_BASE=http://localhost:8082 bash scripts/smoke-rpc.sh
set -euo pipefail

BASE="${API_BASE:-http://localhost:8082}"
CT="Content-Type: application/json"

# 端点清单：path|payload
ENDPOINTS=(
  "antclaw.v1.SystemService/Healthz|{}"
  "antclaw.v1.OptionsService/GetGEX|{\"asset\":\"BTC\"}"
  "antclaw.v1.OptionsService/GetIVSurface|{\"asset\":\"BTC\"}"
  "antclaw.v1.OptionsService/GetOptionsSkew|{\"asset\":\"BTC\"}"
  "antclaw.v1.DeFiService/GetTopProtocols|{\"chain\":\"Ethereum\",\"limit\":2}"
  "antclaw.v1.DeFiService/GetAnalysis|{}"
  "antclaw.v1.SECService/ListFilings|{\"cik\":\"320193\",\"limit\":1}"
  "antclaw.v1.TreasuryService/GetCurve|{}"
  "antclaw.v1.TreasuryService/GetAnalysis|{}"
  "antclaw.v1.OnchainService/GetMetrics|{\"asset\":\"BTC\"}"
  "antclaw.v1.OnchainService/GetAnalysis|{\"asset\":\"BTC\"}"
  "antclaw.v1.FedWatchService/GetFOMCProbabilities|{}"
  "antclaw.v1.SentimentExtrasService/GetCBOEPutCall|{}"
  "antclaw.v1.MacroExtrasService/GetSeries|{\"source\":\"worldbank\",\"series_id\":\"USA/indicator/NY.GDP.MKTP.CD\"}"
  "antclaw.v1.MacroExtrasService/GetSeries|{\"source\":\"imf\",\"series_id\":\"NGDP_RPCH/USA\"}"
  "antclaw.v1.MacroExtrasService/GetSeries|{\"source\":\"ecb\",\"series_id\":\"EXR/D.USD.EUR.SP00.A\"}"
  "antclaw.v1.MacroExtrasService/GetSeries|{\"source\":\"eurostat\",\"series_id\":\"prc_hicp_manr?geo=EU27_2020&coicop=CP00\"}"
  "antclaw.v1.MacroExtrasService/GetSeries|{\"source\":\"snb\",\"series_id\":\"devkua\"}"
  "antclaw.v1.AIService/BuildContext|{\"asset\":\"BTC\",\"scope\":[\"macro\",\"options\"]}"
  "antclaw.v1.SentimentExtrasService/GetMyFXBookPositions|{\"symbol\":\"EURUSD\"}"
  "antclaw.v1.SentimentExtrasService/GetFinvizMetrics|{\"ticker\":\"AAPL\"}"
  "antclaw.v1.RegimeService/GetOverlay|{\"symbol\":\"EURUSD\",\"timeframe\":\"D\",\"contract_code\":\"EUR\"}"
)

fail=0
skip=0
for ep in "${ENDPOINTS[@]}"; do
  path=${ep%%|*}
  data=${ep#*|}
  tmp=$(mktemp)
  code=$(curl -sS -o "$tmp" -w "%{http_code}" -X POST -H "$CT" -d "$data" "$BASE/$path" || echo "000")
  body=$(<"$tmp")
  rm -f "$tmp"
  if [ "$code" = "200" ]; then
    printf "  [%s] %s\n" "$code" "$path"
  elif [ "$path" = "antclaw.v1.RegimeService/GetOverlay" ] && [[ "$body" == *'price_daily'* ]]; then
    # smoke stack price_daily 表为空或缺失 → 视为已知数据空洞跳过
    printf "  [SKIP] %s  (price_daily empty/missing in smoke stack)\n" "$path"
    skip=$((skip + 1))
  elif [ "$code" = "503" ]; then
    # 503 = Connect CodeUnavailable，常见于上游公共 API（DefiLlama / MyFXBook / Finviz / SEC EDGAR / ECB / IMF / 世行 / Eurostat / SNB / Deribit / NY Fed / FRED）
    # 在 CI 偶发限流或网络抖动；不视为代码回退
    printf "  [SKIP] %s  (upstream 503 unavailable)\n" "$path"
    skip=$((skip + 1))
  else
    printf "  [%s] %s  <-- FAIL\n" "$code" "$path"
    fail=$((fail + 1))
  fi
done

if [ "$fail" -eq 0 ]; then
  echo "OK: ${#ENDPOINTS[@]} 个端点中，$(( ${#ENDPOINTS[@]} - skip )) 个通过，$skip 个跳过"
  exit 0
else
  echo "FAIL: $fail 个端点未通过"
  exit 1
fi
