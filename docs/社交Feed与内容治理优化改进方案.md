# 社交 Feed 与内容治理优化改进方案

> 基于 Twitter/X 算法与 Community Notes 研究，结合 AntClaw 项目实际制定

---

## 一、参考系统概述

### 1.1 Twitter/X 推荐系统架构

Twitter 推荐系统由以下核心组件构成：

| 组件类型 | 组件名称 | 职责 |
|---------|---------|------|
| 数据层 | Tweetypie | Tweet 读写服务 |
| | Unified User Actions (UUA) | 用户行为实时流 |
| | User Signal Service (USS) | 用户信号中心化平台 |
| 模型层 | SimClusters | 社区检测与稀疏嵌入 |
| | TwHIN | 用户/推文密集知识图谱嵌入 |
| | Real Graph | 用户间交互预测 |
| | Tweepcred | PageRank 用户声誉计算 |
| | Trust and Safety Models | NSFW/毒性/滥用内容检测 |
| 框架层 | Product Mixer | Feed 构建通用框架 |
| | Home Mixer | 首页 Timeline 服务 |
| | Navi | Rust 高性能模型服务 |
| 推荐流程 | Candidate Sources | 多源候选生成（搜索索引、社交图、Follow推荐） |
| | Light Ranker | 轻量级预排序 |
| | Heavy Ranker | 神经网络重排序 |
| | Visibility Filtering | 可见性过滤（硬过滤、软标签、降权） |

### 1.2 社区标注系统（Community Notes）

- **目标**：让用户为误导性内容添加有帮助的注释
- **评分算法**：基于贡献者评分、note 质量、参与度等维度
- **透明度**：所有 note、评分、贡献者数据公开下载

---

## 二、Feed 推荐系统优化

### 2.1 当前现状分析

**现有架构：**
```
FeedScreen → FeedViewModel → SignalService → SSE Push
```

**问题：**
1. 缺乏多源候选生成机制
2. 没有轻/重排序分层
3. 用户信号体系不完整
4. 缺乏社区发现和嵌入表示

### 2.2 推荐 / 关注 / 最新 / 信号流拆分

#### 2.2.1 四类 Feed 的定位

| Feed 类型 | 用户意图 | 核心信号 | 数据来源 |
|----------|---------|---------|---------|
| **推荐流 (For You)** | 发现感兴趣内容 | 关注图谱 + 兴趣标签 + 互动历史 | SimClusters + UTEG + FRS |
| **关注流 (Following)** | 追踪已关注对象 | 时间线 + 发布频率 | Timeline Service |
| **最新流 (Latest)** | 实时动态 | 时间倒序 | Timeline Service |
| **信号流 (Signals)** | 交易决策支持 | 技术指标 + 交易信号 | Signal Engine |

#### 2.2.2 推荐架构设计

**Pipeline 模式（参考 Product Mixer）：**

```
ProductPipeline (选择哪个 Mixer Pipeline)
    └── MixerPipeline (混合多种异构内容)
            ├── CandidatePipeline (从源获取候选)
            │       ├── SearchIndexCandidateSource (搜索索引)
            │       ├── SocialGraphCandidateSource (社交图)
            │       ├── FollowRecommendationCandidateSource (关注推荐)
            │       └── SignalCandidateSource (信号内容)
            ├── ScoringPipeline (特征填充 + ML 评分)
            └── FilterPipeline (过滤：可见性/去重/多样性)
```

#### 2.2.3 候选源设计

| 候选源 | 触发条件 | 候选数量 | 刷新频率 |
|-------|---------|---------|---------|
| In-Network (关注的人) | 用户有关注 | 50% | 实时 |
| Out-Network (推荐) | 新用户/兴趣匹配 | 50% | 小时级 |
| Search Index | 关键词/话题 | 100 | 分钟级 |
| Follow Recommendations | 冷启动 | 20 | 天级 |
| Signal Engine | 交易者 | 按信号触发 | 实时 |

### 2.3 实施路径

| 阶段 | 内容 | 优先级 |
|-----|------|-------|
| Phase 1 | Feed 类型拆分（四 Tab） | P0 |
| Phase 2 | 候选源抽象接口 | P0 |
| Phase 3 | 轻排序层实现 | P1 |
| Phase 4 | SimClusters 社区嵌入 | P2 |

---

## 三、社交图谱设计

### 3.1 当前现状分析

**现有架构：**
- 用户关系：简单关注列表
- 缺乏：交互图谱、声誉计算、可信度模型

### 3.2 关注体系设计

#### 3.2.1 关注候选生成（参考 FRS - Follow Recommendation Service）

**多源候选策略：**

| 候选源 | 算法 | 描述 |
|-------|------|------|
| Social Graph | 2-hop traversal | 你关注的人关注了谁 |
| Real Graph | ML 分类器 | 预测用户间交互概率 |
| Address Book | 通讯录匹配 | 真实社交关系 |
| PPMI Locale | 地域 + 语言 | 附近用户推荐 |
| Triangular Loops | 三角闭包 | A→B→C→A |

**排序流程：**
```
候选生成 → 特征填充 → ML Ranker → 过滤 → 转换 → 截断
```

#### 3.2.2 作者可信度体系

**多维度评分（参考 Tweepcred + SimClusters）：**

| 维度 | 指标 | 权重 |
|-----|------|------|
| 声誉分 | PageRank | 30% |
| 互动分 | Real Graph 预测 | 20% |
| 内容分 | 信号准确率 | 25% |
| 安全分 | Trust & Safety | 25% |

### 3.3 互动体系设计

#### 3.3.1 互动类型定义

**显式互动（用户主动）：**
- 点赞 (like)
- 评论 (comment)
- 转发 (repost)
- 收藏 (bookmark)

**隐式互动（系统采集）：**
- 浏览 (impression)
- 点击 (click)
- 停留时长 (dwell time)
- 个人主页访问 (profile visit)

#### 3.3.2 互动信号采集（参考 UUA）

**统一用户行为流架构：**

```
客户端/服务端事件 → Kafka → 实时处理 → HDFS / BigQuery / Redis
```

**采集字段：**
```json
{
  "userId": "string",
  "targetType": "post|signal|user|comment",
  "targetId": "string",
  "action": "like|comment|repost|click|view",
  "timestamp": "int64",
  "duration": "int (隐式互动)",
  "deviceInfo": "deviceId, os, appVersion"
}
```

### 3.4 排序影响

**互动信号对排序的影响：**

| 信号类型 | 对推荐的影响 | 对信号排序的影响 |
|---------|-------------|-----------------|
| 高点赞率 | ↑ 内容曝光 | ↑ 信号可信度 |
| 高评论率 | ↑ 内容深度 | ↑ 社区认可 |
| 高转发率 | ↑ 传播广度 | ↑ 影响力 |
| 短停留 | ↓ 内容质量 | ↓ 信号吸引力 |
| 重复点击 | ↑ 兴趣强度 | ↑ 相关性 |

---

## 四、内容治理体系

### 4.1 当前现状分析

**问题：**
- 缺乏内容安全检测
- 没有用户反馈机制
- 无可信标注体系

### 4.2 举报与处理流程

#### 4.2.1 举报类型定义

| 举报类型 | 子类 | 处理优先级 |
|---------|------|----------|
| 垃圾信息 (Spam) | 刷屏、营销、钓鱼 | P1 |
| 滥用 (Abuse) | 骚扰、人身攻击、仇恨言论 | P0 |
| 误导信息 (Misinformation) | 虚假数据、谣言 | P1 |
| 不当内容 (Inappropriate) | NSFW、暴力、血腥 | P0 |
| 版权侵权 (Copyright) | 抄袭、未授权内容 | P2 |

#### 4.2.2 举报 Pipeline（参考 Visibility Filtering）

```
用户举报 → SafetyLabel 标注 → RuleEngine 评估
    → Action: Drop (硬删除) | Label (软标签) | Downrank (降权)
```

**SafetyLevel 分级：**

| 级别 | 场景 | 过滤策略 |
|-----|------|---------|
| Everyone | 公开浏览 | 严格过滤 |
| Followers | 关注者可见 | 中等过滤 |
| Following | 我关注的人 | 宽松过滤 |
| Self | 仅自己 | 无过滤 |

### 4.3 屏蔽与静音

#### 4.3.1 屏蔽（Block）

**屏蔽效果：**
- 双向隔离：屏蔽者与被屏蔽者互相不可见
- 覆盖所有内容：帖子、评论、信号、消息
- 可撤销：临时屏蔽

**屏蔽图谱存储：**
```
UserBlockGraph: UserA → Set<UserB, UserC, ...>
```

#### 4.3.2 静音（Mute）

**静音效果：**
- 单向隐藏：仅屏蔽者看不到
- 不通知被静音者
- 可设置时限：24h / 7d / 永久

**静音类型：**
- 静音用户
- 静音关键词
- 静音话题 (#tag)
- 静音信号类型

### 4.4 可信标注（参考 Community Notes）

#### 4.4.1 标注类型

| 标注类型 | 显示位置 | 作用 |
|---------|---------|------|
| 需要上下文 | 帖子下方 | 补充背景信息 |
| 部分误导 | 帖子下方 + 标签 | 指明错误部分 |
| 争议中 | 帖子下方 + 标签 | 标记存疑 |
| 已被纠正 | 原帖 + 纠正帖 | 链接到真相 |

#### 4.4.2 标注流程

```
用户提交标注申请
    → 评分算法评估（贡献者历史 + Note 质量）
    → 达到阈值 → 公开显示
    → 未达阈值 → 仅自己可见
```

#### 4.4.3 贡献者信誉

| 指标 | 说明 |
|-----|------|
| 评分历史 | 过往标注的采纳率 |
| 争议率 | 被其他用户推翻的比例 |
| 覆盖广度 | 标注的多样性 |
| 时效性 | 是否及时标注 |

### 4.5 低质量内容降权

#### 4.5.1 降权信号（参考 Trust and Safety Models）

**检测模型：**
- pNSFWMedia: NSFW 图片检测
- pNSFWText: NSFW 文本检测
- pToxicity: 毒性内容检测
- pAbuse: 滥用行为检测

#### 4.5.2 降权策略

| 降权因子 | 降权幅度 | 恢复条件 |
|---------|---------|---------|
| 垃圾信息 | -80% | 无 |
| 轻度毒性 | -50% | 7 天无违规 |
| 高争议 | -30% | 争议解除 |
| 低互动 | -20% | 互动率提升 |
| 旧内容 | -10%/天 | 无 |

---

## 五、通知体系

### 5.1 当前现状分析

**现有通知类型：**
- 信号提醒 (Signal Alert)
- 价格提醒 (Price Alert)
- 系统通知 (System Notification)

**问题：**
- 没有社交通知
- 没有互动通知
- 缺乏通知排序

### 5.2 通知分类体系

#### 5.2.1 四大通知类型

| 类型 | 触发条件 | 优先级 | 实时性 |
|-----|---------|-------|-------|
| **互动通知** | 点赞/评论/转发/关注 | P1 | 实时 |
| **交易信号通知** | 策略触发/止损/止盈 | P0 | 实时 |
| **系统通知** | 账户变动/安全提醒 | P0 | 实时 |
| **社区通知** | 标注更新/热门话题 | P2 | 小时级 |

#### 5.2.2 互动通知详情

| 子类型 | 触发 | 内容模板 |
|-------|------|---------|
| 新粉丝 | 关注操作 | "{user} 关注了你" |
| 点赞 | 点赞你的帖子 | "{user} 赞了你的 {type}" |
| 评论 | 评论你的帖子 | "{user} 评论了你的 {type}" |
| 转发 | 转发你的帖子 | "{user} 转发了你的 {type}" |
| 提及 | @提及你 | "{user} 在 {type} 中提及了你" |
| 回复 | 回复你的评论 | "{user} 回复了你的评论" |

#### 5.2.3 通知排序（参考 Pushservice）

**通知 Ranking 流程：**

```
Target 构建 → 候选获取 → 特征填充 → Light Ranker → Heavy Ranker → Take Step → 发送
```

**Ranking 特征：**

| 特征类型 | 特征名 | 来源 |
|---------|-------|------|
| 用户特征 | 互动概率 | Real Graph |
| | 通知打开率 | Historical Data |
| 内容特征 | 质量分 | Trust & Safety |
| | 新鲜度 | Timestamp |
| | 热门度 | Engagement Count |
| 上下文 | 距上次通知 | Time Delta |
| | 通知频率上限 | Rate Limit |

### 5.3 通知控制

#### 5.3.1 用户通知偏好

**渠道控制：**
- 推送通知 (Push)
- 应用内通知 (In-App)
- 邮件通知 (Email)
- SMS 通知

**类型控制：**
```json
{
  "likes": { "push": true, "inApp": true, "email": false },
  "comments": { "push": true, "inApp": true, "email": true },
  "newFollowers": { "push": true, "inApp": true, "email": false },
  "signals": { "push": true, "inApp": true, "email": true }
}
```

#### 5.3.2 通知频率控制

| 场景 | 上限 |
|-----|------|
| 单用户单类型 | 10/小时 |
| 单用户全局 | 50/小时 |
| 信号提醒 | 无限制（P0） |

---

## 六、Profile 主页优化

### 6.1 当前现状分析

**现有 Profile 内容：**
- displayName
- bio
- 关注数/粉丝数
- 交易指标（胜率/盈亏比）

**问题：**
- 缺乏身份验证信息
- 缺乏内容 Tab
- 缺乏社交证明

### 6.2 Profile 结构设计

#### 6.2.1 Header 区域

```text
┌─────────────────────────────────────────┐
│  [头像]                                 │
│  显示名称    ✓(认证)  🔒(私密)           │
│  @用户名 / codeId                        │
│                                         │
│  生物简介                                │
│                                         │
│  📅 注册时间  📍 地区                    │
│                                         │
│  [关注] [交易]    粉丝 123  关注 456     │
└─────────────────────────────────────────┘
```

#### 6.2.2 交易员信任指标

| 指标 | 说明 | 可信度来源 |
|-----|------|----------|
| 胜率 | 盈利交易/总交易 | MT 数据 |
| 盈亏比 | 平均盈利/平均亏损 | MT 数据 |
| 夏普比率 | 风险调整收益 | 计算得出 |
| 最大回撤 | 历史最大亏损 | MT 数据 |
| MT 认证 | MT4/MT5 绑定验证 | 账户验证 |

#### 6.2.3 Tab 体系

| Tab | 内容 | 数据源 |
|-----|------|-------|
| 帖子 | 用户发布的帖子 | FeedService/ListUserPosts |
| 回复 | 用户的评论/回复 | FeedService/ListUserReplies |
| 信号 | 用户的交易信号 | SignalService/ListUserSignals |
| 账户 | 绑定的交易账户 | MT Service |

### 6.3 社交证明设计

#### 6.3.1 作者认证体系

| 认证类型 | 图标 | 获取条件 |
|---------|------|---------|
| 身份认证 | ✓ 蓝 | 官方验证 |
| 交易员认证 | ✓ 金 | MT 数据达标 |
| 社区认证 | ✓ 银 | Community Notes 贡献 |

#### 6.3.2 Social Context

**显示方式：**
- "被 {X} 等 {N} 人关注"
- "在 {Community} 中知名"
- "与你有 {N} 个共同关注"

---

## 七、实施计划

### 7.1 优先级矩阵

| 模块 | 影响范围 | 实现复杂度 | 优先级 |
|-----|---------|----------|-------|
| Feed 类型拆分 | 高 | 中 | P0 |
| 通知体系重构 | 高 | 高 | P0 |
| 用户信号采集 | 高 | 中 | P0 |
| 内容安全检测 | 高 | 高 | P1 |
| 关注推荐系统 | 中 | 高 | P1 |
| Community Notes | 中 | 高 | P2 |

### 7.2 分阶段实施

#### Phase 1：基础设施（1-2 周）

| 任务 | 交付物 |
|-----|-------|
| 用户信号采集框架 | UnifiedUserActions Service |
| 行为数据存储 | Redis + BigQuery |
| 基础举报功能 | ReportService |

#### Phase 2：Feed 优化（2-3 周）

| 任务 | 交付物 |
|-----|-------|
| 四类 Feed 拆分 | FeedScreen v2 |
| 候选源抽象接口 | CandidateSource Interface |
| 轻排序实现 | LightRanker Service |

#### Phase 3：通知与治理（2-3 周）

| 任务 | 交付物 |
|-----|-------|
| 互动通知 | NotificationService v2 |
| 通知排序 | PushRanker |
| 内容安全检测 | SafetyFilter |
| 屏蔽/静音 | BlockMuteService |

#### Phase 4：高级特性（3-4 周）

| 任务 | 交付物 |
|-----|-------|
| SimClusters 社区嵌入 | CommunityDetection |
| Real Graph 交互预测 | InteractionPredictor |
| Community Notes | CrowdsourcedAnnotations |
| Profile 社交证明 | SocialProof UI |

---

## 八、技术依赖

### 8.1 服务端依赖

| 组件 | 语言 | 用途 |
|-----|------|------|
| Kafka | - | 用户行为流 |
| Redis | - | 实时特征存储 |
| BigQuery | SQL | 离线特征计算 |
| Go Service | Go | 微服务实现 |

### 8.2 数据模型

**核心实体：**
- User: 用户基本信息
- Post: 帖子/信号
- Interaction: 用户互动
- TrustScore: 可信度评分
- SafetyLabel: 安全标注

### 8.3 API 扩展

| Proto 定义 | 方法 | 说明 |
|-----------|------|------|
| FeedService | ListForYou | 推荐流 |
| | ListFollowing | 关注流 |
| | ListLatest | 最新流 |
| NotificationService | GetNotifications | 获取通知 |
| | UpdatePreferences | 更新偏好 |
| ModerationService | ReportContent | 举报内容 |
| | BlockUser / UnblockUser | 屏蔽用户 |
| | MuteUser / UnmuteUser | 静音用户 |
| TrustService | GetTrustScore | 获取可信度分 |
| | GetAnnotations | 获取内容标注 |

---

## 九、现有项目适配

### 9.1 已有组件映射

| 现有组件 | 可复用模块 | 适配方案 |
|---------|----------|--------|
| FeedService | 候选获取层 | 扩展 CandidateSource |
| SignalService | 信号内容生成 | 复用为 SignalCandidateSource |
| AlertService | 价格提醒 | 复用为 SignalNotification |
| PresenceService | 在线状态 | 复用为 UserSignal 一部分 |

### 9.2 新增组件清单

| 新增服务 | 职责 | 优先级 |
|---------|------|-------|
| UserActionCollector | 统一行为采集 | P0 |
| NotificationRanker | 通知排序 | P0 |
| ContentSafetyFilter | 内容安全过滤 | P1 |
| FollowRecommendationService | 关注推荐 | P1 |
| TrustScoreService | 可信度评分 | P1 |
| CommunityAnnotationService | 社区标注 | P2 |

---

## 十、风险与缓解

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 数据隐私合规 | 高 | 明确 consent，遵循最小采集原则 |
| 通知骚扰 | 中 | 严格频率限制，用户可完全关闭 |
| 推荐同质化 | 中 | 引入多样性约束 |
| ML 模型偏差 | 中 | 定期审计，增加人工兜底 |
| 恶意举报泛滥 | 中 | 举报成功率与信誉关联 |

---

*文档版本：v1.0*
*创建日期：2026-05-16*
*参考来源：Twitter/X The Algorithm, Community Notes*
