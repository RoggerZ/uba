# GitHub 账号和仓库组织应该怎么规划

## Q：我想用一个空账号专门维护 SimpleTrack，但现在 `RoggerZ` 改不了，怎么办？

A：如果 `RoggerZ` 这个用户名不能修改或不适合作为长期品牌名，建议不要把项目品牌强绑定到个人用户名。更稳的做法是创建一个 **GitHub Organization**，用组织来承载公开产品仓库。

例如：

- 个人账号：继续保留，放自己的实验、私有草稿、随手项目。
- 组织账号：用于公开产品，例如 `simpletrack`、`apptrack`、`analytics-core`。

这样外部看到的是一个专业组织，而不是你的个人杂物间。

## Q：后面做 AppTrack，还需要再新建一个账号吗？

A：不需要。更推荐一个统一组织承载多个产品：

- `analytics-core`：通用分析核心。
- `simpletrack`：Web/SaaS 行为分析产品。
- `apptrack`：App 行为分析产品。

这样三个仓库看起来是同一套产品线，外部品牌更专业，也方便复用 `analytics-core`。

## Q：我的其他私有仓库也想放一起，可以吗？

A：可以，但要分清“外部可见”和“内部随便”。

建议结构：

| 类型 | 可见性 | 放哪里 |
| --- | --- | --- |
| 对外产品仓库 | Public | 统一 GitHub Organization |
| 私有产品源码 | Private | 同一个 Organization 或个人账号均可 |
| 个人实验、随手脚本 | Private | 个人账号更合适 |
| 文档、官网、SDK | Public 或 Private | 看是否需要对外展示 |

你的目标“外部只可见 SimpleTrack、AppTrack，并且很专业；自己的东西自己可见，可以随便一点”很合理。实现方式就是：公开仓库放组织里，个人草稿保持 private。

## Q：Organization 名字应该怎么选？

A：不要用太具体的单产品名，除非只做一个产品。你现在已经计划 SimpleTrack、AppTrack、analytics-core，建议组织名偏中性、产品线化。

命名原则：

- 不带个人昵称。
- 不带某个具体产品名，避免以后扩展尴尬。
- 简短、好读、像公司或产品工作室。
- 公开仓库 README、头像、简介保持专业。

## Q：`analytics-core` 应该放在哪里？

A：建议放在统一组织下，作为多个产品共用的核心仓库。后续关系是：

- xwl_bi 逐步被 `analytics-core` 反向支撑。
- SimpleTrack 引用 `analytics-core`。
- AppTrack 也可以引用 `analytics-core`。

所以 `analytics-core` 不应该放在某个产品专属账号下，也不应该命名成 `simpletrack-core` 或 `xwl-core`。
