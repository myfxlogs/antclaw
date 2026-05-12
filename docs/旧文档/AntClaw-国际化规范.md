# AntClaw · 国际化规范（i18n）

> 本文是重构方案 §十二（国际化）与决策 #18 的实现细则。任何与本文冲突的前后端实现必须被拒绝合并。

## 一、原则

1. **全链路 i18n**：后端错误、通知、邮件、推送、审计 summary、前端界面、枚举显示文案，一律走 i18n。
2. **首发语种**：`zh-CN`（简体中文）、`en-US`（英文）。其他语种（`ja-JP`、`ko-KR`、`zh-TW`）为**预留**。
3. **兜底（fallback）**：
   - 用户端（Web / 移动端）：**`en-US`**。
   - 管理端（Admin）：**`zh-CN`**。
4. **key 优先**：后端只返回 `message_key` 与参数，**不拼接人话**；人话在前端或通知层按用户 locale 渲染。
5. **不在代码中硬编码可见字符串**（日志除外，日志固定 `en-US`）。

## 二、locale 协商

### 2.1 识别顺序

1. 用户资料 `users.locale`（已登录）。
2. Cookie `antclaw_locale`（一次性 API 切换写入）。
3. HTTP `Accept-Language`（BCP-47，取 q 最高且被支持者）。
4. 查询参数 `?locale=<tag>`（仅白名单内）。
5. 兜底：客户端类型决定（用户端 `en-US`；管理端 `zh-CN`）。

### 2.2 白名单

```
supported = {"zh-CN", "en-US"}
reserved  = {"ja-JP", "ko-KR", "zh-TW"}   // 结构支持但资源未齐；回落到 fallback
```

### 2.3 传输

- 请求头：`Accept-Language: zh-CN,en-US;q=0.8`。
- 响应头：`Content-Language: <最终选中 tag>`。
- 所有 RPC 响应在 `ErrorDetail` 中附 `locale` 字段，前端可做日志关联。

## 三、key 命名与资源结构

### 3.1 命名规范

```
<域>.<模块>.<用途>.<标识>
```

- 小写、`.` 分隔、`_` 用于词内分隔；禁止驼峰。
- 错误消息：`err.<domain>.<code>`；例：`err.auth.invalid_credentials`。
- UI 文案：`ui.<page>.<element>`；例：`ui.dashboard.title`。
- 通知：`notify.<channel>.<template>`；例：`notify.inapp.cot_alert_v1`。
- 邮件：`mail.<template>.<section>`；例：`mail.password_reset.subject` / `...body`。
- 枚举：`enum.<type>.<value>`；例：`enum.signal_severity.high`。

### 3.2 ICU MessageFormat

所有可参数化 key 使用 ICU：

```
err.auth.login_cooldown = "登录已锁定，请在 {seconds, plural, one {# 秒} other {# 秒}} 后重试。"
ui.alerts.unread_count  = "{count, plural, =0 {无未读} one {# 条未读} other {# 条未读}}"
mail.password_reset.body = "您好 {name}，..."
```

### 3.3 文件布局

```
frontend/packages/i18n/
├── zh-CN/
│   ├── common.json
│   ├── ui.json
│   ├── err.json
│   ├── enum.json
│   ├── notify.json
│   └── mail.json
├── en-US/
│   └── ... （与 zh-CN 同结构同 key）
└── index.ts    # 导出 i18n 实例，供 Web/Admin/移动端共享
```

后端镜像：

```
backend/internal/i18n/messages/
├── zh-CN/...
└── en-US/...
```

- 后端/前端资源以 `scripts/i18n_sync.ts` 双向校验：**同 key 必须在两端都存在**；CI 失败即阻断合并。

## 四、后端实现

### 4.1 库

- Go：`github.com/nicksnyder/go-i18n/v2`（ICU 子集）+ `golang.org/x/text/language`（匹配）。
- 资源加载：启动期 `Walk` `messages/<locale>/*.json` 全部装载；热更新由管理员通过 `POST /admin/i18n/reload` 触发（仅内存重建，不重启）。

### 4.2 使用位置

- **错误响应**：handler 层返回 `connect.NewError(code, err)` + `ErrorDetail{ message_key, args }`；**不**在服务端渲染成文字。
- **通知（站内信）**：`notifications` 表只存 `message_key` 与 `args_json`；推送/展示时由对应客户端本地渲染。
- **邮件**：worker 按用户 `locale` 本地渲染；模板按 `mail.<template>.<locale>.html` 存放。
- **管理端审计 summary**：以 `zh-CN` 硬渲染入表的 `summary` 列（便于全文检索），同时保留 `message_key` + `args_json`。

### 4.3 动态业务字段（`title_i18n`）

- 多语字段用 JSONB：`title_i18n JSONB NOT NULL DEFAULT '{}'::jsonb`。
- 约束：`title_i18n ? 'zh-CN'`（必含 `zh-CN`）；其他 locale 可选。
- 读取回退：`title_i18n ->> <user_locale>` 为空 → `title_i18n ->> <fallback>` → `title_i18n ->> 'zh-CN'`。
- SQL 辅助函数 `i18n_pick(j jsonb, locale text, fallback text) RETURNS text` 集中封装。

### 4.4 时区与时间格式

- 存库一律 `timestamptz`（UTC）；
- 响应序列化一律 RFC 3339；
- 前端按用户 `timezone` + `locale` 渲染（Intl.DateTimeFormat）；后端不返回已格式化字符串。
- 数字/货币同理：后端返 `Money{currency, units, nanos}`，前端按 locale 渲染。

## 五、前端实现

### 5.1 库

- `react-i18next` + `i18next-icu` + `i18next-http-backend`（Web/Admin）。
- React Native：`i18next` + `react-i18next`（平台检测 locale）。
- 资源打包：构建期把 `frontend/packages/i18n/<locale>/*.json` 打进 bundle；**允许**按 locale 拆 chunk 延迟加载。

### 5.2 API

```ts
const { t } = useTranslation(['common', 'ui', 'err']);
t('ui.dashboard.title');
t('ui.alerts.unread_count', { count: 3 });
t('err.auth.invalid_credentials');  // 兜底: 后端 ErrorDetail.message_key 映射
```

### 5.3 错误渲染

- 所有 RPC catch 走 `renderRpcError(err, t)`：
  - 优先 `t(err.details.message_key, err.details.args)`。
  - 缺 key → `t('err.common.unknown')`。
  - **禁止**直接显示后端 `err.message`（可能泄漏栈/英文）。

### 5.4 用户切换 locale

- `UserService.UpdateSettings({ locale })` → 同时写 cookie `antclaw_locale`。
- 切换后强制 `i18n.changeLanguage(newLocale)` 并 **刷新当前路由**，不重登。

## 六、管理端专用

- 管理端 `frontend/admin/` **不读** 用户端 locale；独立走 `admin_locale` cookie。
- 兜底 `zh-CN`；即使管理员 `users.locale = en-US`，Admin 界面仍默认 `zh-CN`（可在 Admin 内单独切换到 `en-US`）。
- `i18n_strings` 管理页面：
  - 表：`i18n_strings(key, locale, value, updated_by, updated_at)`；
  - 缺失键看板：对比两套 locale 的 key 差集；
  - 导入/导出 JSON；
  - 编辑后 **写审计** + 触发内存 reload。

## 七、CI 检查（`i18n-check` job）

1. 扫描 `frontend/packages/i18n/<locale>/*.json` 所有 key 集合。
2. 每个 locale 必须拥有**完全一致**的 key 集合（首发支持语种之间）。
3. 扫描 `grep -rnE "t\(['\"][a-z0-9_.]+['\"]"` 使用的 key，与资源文件交叉校验：
   - 资源中定义但未使用 → 警告（不阻断）。
   - 代码中使用但资源未定义 → **阻断**。
4. 扫描 `frontend/` 代码中的中文字符：除 `*.i18n-allow.*` 白名单和注释外，出现中文字符串字面量 → **阻断**。
5. 扫描后端代码中的 `errors.New("...")` / `fmt.Errorf("...中文...")`：禁止中文硬编码，日志允许英文 → 警告。
6. ICU 语法校验：`intl-messageformat-parser` 解析所有值，语法错误阻断。

## 八、枚举本地化

- proto 枚举保留原值；前端显示走 `enum.<type>.<value>`。
- 示例：

```json
{
  "enum.signal_severity.low": "低",
  "enum.signal_severity.medium": "中",
  "enum.signal_severity.high": "高",
  "enum.signal_severity.critical": "严重"
}
```

- 后端不返回中英文本，只返回枚举字符串。

## 九、测试要求

- **单测**：`i18n_pick` 的回退逻辑全覆盖（用户 locale 存在 / 不存在 / JSON 为空 / 异常 locale）。
- **快照测试**：常见页面两种 locale 快照对比，key 未找到即失败。
- **契约测试**：错误返回 `message_key` 必须在资源库中；用 CI fixture 固化。
- **端到端**：切换 locale 后关键页面立即重渲染。
- **负向**：未知 locale（`xx-YY`）应回落到 fallback 且不报错。

## 十、验收清单（对照任务卡 P8）

- [ ] `zh-CN`、`en-US` 两套资源文件齐全、key 集合完全一致。
- [ ] Admin 默认 `zh-CN`，用户端默认 `en-US`（兜底）。
- [ ] `i18n-check` CI job 全部通过。
- [ ] 后端所有错误响应都带 `message_key`；前端全部通过 `renderRpcError` 映射。
- [ ] `title_i18n` JSONB 字段的读取回退 SQL 函数有集成测试。
- [ ] 邮件模板两语种齐全，worker 渲染单测覆盖。
- [ ] 管理员可在 Admin 控制台编辑 `i18n_strings`，热更新生效、写审计。

## 十一、已决事项（2026-04-24）

- **I1 · `zh-TW` 机翻占位**：**不启用**；首发仅 `zh-CN` + `en-US`；`zh-TW` / `ja-JP` / `ko-KR` 资源结构保留但不生成。
- **I2 · 货币显示**：**跟随业务**；后端按业务语义返回 `Money{currency}`，前端按字面渲染，不做 locale 换算（即不做汇率折算）。
