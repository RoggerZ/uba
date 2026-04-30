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

## Q：Litlyx 是怎么处理的？

A：Litlyx 更强调 workspace 和 Product 视角。它的接入脚本里会出现 workspace 标识，用户通过 Raw Events 和 Product 验证数据是否进入。

可以理解成：

```text
workspace
  └── product / raw events
```

Litlyx 的表达适合新手接入和首价值，但 `analytics-core` 仍然应该抽象成更通用的 `tenant / project / source`。

## Q：当前建议怎么定？

A：`analytics-core` 采用：

- `tenant_id`：最外层隔离边界。
- `project_id`：分析项目边界。
- `source_id`：事件来源边界。
- `source_type`：标识 source 是 `web`、`ios`、`android`、`server` 等。

这样既能覆盖 SimpleTrack，也能覆盖 AppTrack，还能反向支撑 xwl_bi。
