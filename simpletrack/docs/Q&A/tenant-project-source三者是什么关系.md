# tenant、project、source 三者是什么关系

## Q：`tenant_id`、`project_id`、`source_id` 是并列关系还是包含关系？

A：它们不是并列关系，而是包含关系：

```text
tenant
  └── project
        └── source
```

推荐解释：

- `tenant_id`：租户或客户边界，通常对应一个组织、账号主体或业务客户。
- `project_id`：租户下面的分析项目，用来组织一组相关数据源和报表。
- `source_id`：真正产生事件的数据源，可以是网站、App、服务端、小游戏、小程序等。

## Q：为什么不用 `workspace_id / site_id`？

A：因为 `analytics-core` 要保持业务无关。`workspace` 和 `site` 更像 SimpleTrack 产品层概念，不适合写死进核心库。

如果未来 AppTrack 接入，数据源可能不是 `site`，而是：

- iOS App。
- Android App。
- 小程序。
- 后端服务。

所以核心库使用更通用的 `source_id`。

## Q：SimpleTrack 怎么映射？

A：SimpleTrack 可以这样映射：

| SimpleTrack 产品层 | analytics-core 核心层 |
| --- | --- |
| Workspace / Organization | `tenant_id` |
| Project / Website Group | `project_id` |
| Website / Domain | `source_id` |

P1 如果暂时没有复杂 project，也可以先让 `project_id` 等于 `tenant_id` 或由系统自动创建默认 project，但事件协议里保留这个层级。

更准确地说，SimpleTrack 的第一版很可能是：

```text
workspace
  └── default project 或 website group
        └── website / domain / tracking source
```

如果一个 workspace 里只有一个网站，`project_id` 和 `source_id` 可以是一对一；如果后续一个客户有多个站点、多个环境或 Web + Server 多端采集，`project` 就能把这些 source 组织到同一个分析项目里。

## Q：AppTrack 怎么映射？

A：AppTrack 可以这样映射：

| AppTrack 产品层 | analytics-core 核心层 |
| --- | --- |
| Organization / Account | `tenant_id` |
| App Project | `project_id` |
| iOS App / Android App / 小程序 | `source_id` |

这就是为什么 `analytics-core` 不应该用 `site_id` 作为核心字段。

## Q：Umami 是怎么处理的？

A：Umami 的核心对象更偏 Website。典型关系是：

```text
team / user
  └── website
        └── events / sessions / reports
```

Umami 的 API 和文档里大量围绕 `websiteId` 展开，例如 Realtime、Events、Sessions 都是围绕某个 website 查询。这对 Web Analytics 很直接，但对 AppTrack 这种多端产品不够通用。

对应到 `analytics-core`：

| Umami | analytics-core | 说明 |
| --- | --- | --- |
| team / user | `tenant_id` | 最外层归属和隔离边界 |
| website | `project_id` | Umami 的 website 通常直接就是一个分析项目 |
| 无独立 source 层 | `source_id` | `analytics-core` 多出来的通用数据源层；映射 Umami 时可以创建 default source，或让 `source_id = website_id` |

所以，是的：相对 Umami，`analytics-core` 多了一个显式 `source` 维度。这个维度是为了后续支持同一个分析项目下的多个来源，例如 web、server、iOS、Android、小程序或多个域名。对纯 Web Analytics 场景，它可以先保持一对一，不增加产品复杂度。

## Q：Litlyx 是怎么处理的？

A：Litlyx 更强调 workspace 和 Product 视角。它的接入脚本里会出现 workspace 标识，用户通过 Raw Events 和 Product 验证数据是否进入。

可以理解成：

```text
workspace
  └── product / raw events
```

Litlyx 的表达适合新手接入和首价值，但 `analytics-core` 仍然应该抽象成更通用的 `tenant / project / source`。

对应到 `analytics-core`：

| Litlyx | analytics-core | 说明 |
| --- | --- | --- |
| workspace | `tenant_id` | 租户、账号或团队边界 |
| product | `project_id` | 产品或分析项目 |
| 接入 SDK / 真实产生数据的端 | `source_id` | Litlyx 产品层不一定显式暴露，但核心层需要保留 |

所以，相对 Litlyx，`analytics-core` 也多了一个可显式建模的 `source` 维度。P1 可以让一个 product 默认只有一个 source；后续如果同一个 product 同时接 Web SDK、Server SDK、移动端 SDK，就不用重做核心事件模型。

## Q：xwl_bi 怎么映射？

A：xwl_bi 主要按 `appid` 分表，是一个维度。迁到 `analytics-core` 时不要机械把 `appid` 当成全部三层，而是按迁移粒度补齐默认层级。

推荐映射：

| xwl_bi | analytics-core | 说明 |
| --- | --- | --- |
| 原系统或客户边界 | `tenant_id` | 如果旧系统没有租户概念，迁移时创建默认 tenant |
| app 或分析项目 | `project_id` | 如果 `appid` 在旧系统代表一个完整应用/产品，可以映射到 project |
| appid / table_id / 数据源 | `source_id` | 如果 `appid` 更像事件来源或物理分表键，则映射到 source |

实践上可以先采用：

```text
default tenant
  └── project derived from appid
        └── source derived from appid
```

也就是让 `project_id` 和 `source_id` 初期一对一。等 xwl_bi 后续被 `analytics-core` 反向支撑时，再逐步把旧的 appid 分表逻辑收敛到 ClickHouse adapter 的物理表策略里。

## Q：`source_id` 是不是一定要在产品界面展示？

A：不一定。`source_id` 是核心层概念，不一定要完整暴露给 SimpleTrack 用户。

P1 产品界面可以只让用户看到 Website / Product / Project。系统内部自动创建默认 source，并生成 write key 或 tracking source。等到需要多端接入时，再把 source 作为“数据源”“接入源”或“SDK 来源”展示出来。

## Q：当前建议怎么定？

A：`analytics-core` 采用：

- `tenant_id`：最外层隔离边界。
- `project_id`：分析项目边界。
- `source_id`：事件来源边界。
- `source_type`：标识 source 是 `web`、`ios`、`android`、`server` 等。

这样既能覆盖 SimpleTrack，也能覆盖 AppTrack，还能反向支撑 xwl_bi。

参考：

- Umami Websites API：`https://docs.umami.is/docs/api/websites`
- Umami Teams：`https://umami.is/docs/teams`
- Litlyx Easy Setup：`https://docs.litlyx.com/universal`
- Litlyx Custom Events：`https://docs.litlyx.com/custom-events`
