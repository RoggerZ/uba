# Supastarter 内置 PostgreSQL 是什么作用

## Q：模板内置的 PostgreSQL 是什么作用？

A：Supastarter 内置的 PostgreSQL 主要是 **SaaS 控制面数据库**，不是 SimpleTrack 的行为分析事件库。

它负责存储“谁能使用系统、属于哪个组织、买了什么套餐、系统如何配置”这类低频事务数据。以当前 `src/simpletrack-saas` 的 Prisma schema 为例，PostgreSQL 会存：

- 用户：`User`。
- 登录会话：`Session`。
- 登录账号和 OAuth 绑定：`Account`。
- 邮箱验证、密码重置、magic link 等校验记录：`Verification`。
- passkey、two-factor 等认证安全数据：`Passkey`、`TwoFactor`。
- 组织、成员、邀请：`Organization`、`Member`、`Invitation`。
- 支付和订阅记录：`Purchase`。
- 站内通知和通知偏好：`Notification`、`UserNotificationPreference`。

## Q：SimpleTrack 后续可以把什么放进 PostgreSQL？

A：可以放低频业务配置和控制面状态，例如：

- Workspace / Organization。
- Site / Website / Source 元数据。
- 安装 snippet、接入状态、source 是否启用。
- Goal 定义。
- 当前套餐、source 数量限制、subscription gate 状态。
- onboarding 进度。
- Admin/support 操作需要的用户、组织和订阅信息。

这些数据更适合 PostgreSQL，因为它们需要事务、一致性、关系查询和后台管理，不是高吞吐事件明细。

## Q：那事件数据放在哪里？

A：事件数据不要放 PostgreSQL 作为主存储。

SimpleTrack 的 pageview、自定义事件、实时写入、漏斗、留存、路径、分群、归因等分析数据面，应由 `analytics-core` 承接：

- Redis Stream 或 Kafka：事件流、异步消费、ack、重试。
- ClickHouse：事件明细、聚合查询、实时和历史分析。
- MySQL / PostgreSQL：只在需要时存元数据或控制面配置，不承担大规模事件明细。

一句话：**PostgreSQL 管控制面，ClickHouse 和事件流管分析数据面。**

## Q：本机现在没有 PostgreSQL，会阻塞当前工作吗？

A：不阻塞 marketing、docs、mail-preview 和 TypeScript type-check。

需要真实数据库的场景主要是：

- 本地注册、登录、会话。
- 创建组织、成员邀请。
- 跑 authenticated SaaS 页面真实流程。
- 测试支付订阅写入和 purchase 查询。
- 将 SimpleTrack 的 site/source/goal 元数据真正落库。

## Q：如果本机需要 PostgreSQL，怎么启动？

A：`src/simpletrack-saas` 已有 Docker Compose 配置，当前已改成 SimpleTrack 语义：

```powershell
cd C:\Users\1\Documents\git\uba\src\simpletrack-saas
docker compose up -d postgres
```

对应连接串：

```powershell
$env:DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/simpletrack"
```

注意：如果本机以前已经用旧模板配置创建过 `supastarter` 数据卷，Docker 不会因为修改 `POSTGRES_DB` 自动重建数据库。当前用户说明本机还没有这个数据库，所以可以直接按新配置启动；如果未来遇到旧数据卷冲突，再单独记录到 `docs/开发环境卡壳问题记录.md`。
