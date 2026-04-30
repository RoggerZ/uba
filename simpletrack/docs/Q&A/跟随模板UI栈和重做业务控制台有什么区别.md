# 跟随模板 UI 栈和重做业务控制台有什么区别

## Q：跟随 Supastarter UI 栈是什么意思？

A：意思是继续使用 Supastarter 已经选好的技术和组件基础，例如 Tailwind、shadcn/ui、布局、按钮、表单、弹窗、导航、设置页结构等。

这样做的重点是：**不另起一套 UI 基础设施**。

优点：

- 开发快。
- 和模板的 auth、settings、billing、admin 页面风格一致。
- 少维护一套组件系统。
- 更容易跟随模板更新。

## Q：在模板 shell 内重做业务控制台是什么意思？

A：意思是保留 Supastarter 的应用壳，例如登录、组织、侧边栏、顶栏、权限、路由、布局，但 SimpleTrack 的核心分析页面重新按企业分析控制台设计。

例如：

- Realtime 页面自己设计信息密度。
- Events 表格自己设计列、筛选、明细抽屉。
- Website settings 自己设计安装和接入验收流程。
- Goal、Funnels、Segments 等分析页面按 SimpleTrack 的产品逻辑组织。

## Q：这两个是二选一吗？

A：不是绝对二选一。推荐策略是：

1. **底层 UI 栈跟随 Supastarter**。
2. **业务页面在 Supastarter shell 内重做**。

也就是不造组件基础设施，但业务体验不能被模板默认页面牵着走。

## Q：产品官网 / docs 怎么做？

A：先使用 Supastarter 自带的 `apps/marketing` 和 `apps/docs` 看效果是否满足。

- 如果效果好，直接使用。
- 如果信息架构或视觉表达不够 SimpleTrack 专业，再轻量定制。
- 不建议 P1 另起一套独立官网框架。

## Q：当前决策是什么？

A：当前决策是：

- 商业控制面和 app shell：使用 Supastarter。
- 产品官网 / docs：先使用 Supastarter 的 marketing/docs app。
- 分析业务控制台：在 Supastarter shell 内按 SimpleTrack 需求重做核心页面。
