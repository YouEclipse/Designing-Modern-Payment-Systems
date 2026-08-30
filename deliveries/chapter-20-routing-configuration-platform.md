# 第 20 章《路由配置平台》交付记录

## 章节契约

- 核心命题：配置包、发布与启用是三种不同事实；审批固定不可变内容，运行实例一次只启用完整快照。
- 演进线：环境变量与文件 → 跨制品依赖闭包 → 审批状态机 → 原子启用、最近一个已知可用版本与有限紧急停用开关。
- 本章代码基线：路由服务 v3 已严格编译完整的成本、性能与评分数据平面配置；独立的路由配置控制面尚未实现。
- 后续范围：模型制品作为一种受治理制品留给第 21 章。

## 能力状态

- `implemented`：固定实现中的严格 v3 启动配置、完整优化依赖快照、与调用方输入深拷贝隔离和确定重放。
- `teaching_model`：正文中的配置包、发布与启用状态机，以及原子运行时切换 Go 模型。
- `target_contract`：独立路由配置服务、双人审批、签名、分发发件箱、可信缓存、监视器、启用确认与跨区发布。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库链接需授权；不改变可见性。
- 本章没有虚构 `services/routing-config`、审批数据库或监视器运行时。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 运行请求不能提交自己的策略或制品 | 路由服务 OpenAPI 0.4；`services/router/internal/httpapi` 严格读取器 | 严格 JSON 与 OpenAPI 契约测试；v3 黄金夹具 |
| `implemented` | v3 启动快照必须完整，风险、成本、性能与评分依赖共同生效 | `services/router/cmd/server/v3_config.go`；`decision/optimization.go` | `TestDevelopmentV3IsRiskGatedAndUsesExactLocalConnections`；`TestV2AndV3ConfigurationReadersCannotBeConfused` |
| `implemented` | 同一引擎的制品快照可确定重放且不受别名修改 | `services/router/internal/{cost,performance,scoring,decision}` | 规范归属与并发重放测试；`TestOptimizationSnapshotDigestOrderOwnershipAndConcurrentReplay` |
| `teaching_model` | 配置包、发布与启用生命周期仍是教学合同 | 本章 `configrelease` 与 `runtimeconfig` Go 包 | `scripts/go-examples/tests/chapter20` 编译、竞态与静态检查；状态机与原子切换测试 |

## 图表与书稿验证

- Mermaid 覆盖归属、审批状态机、发布与启用分离及完整运行时切换。
- 本章 Go 示例以第 20 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；发布代际与原子性由状态/时序图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/configuration-platform.md`、`scripts/verify-mermaid.sh part-5-global-routing/configuration-platform.md` 与 `git diff --check`；这些门禁只证明发布与运行时配置教学模型。
- 当前运行时主张只能引用交付清单中列出的 v3 启动快照代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；严格启动解析不等于控制平面或热发布。

## 未实现与能力边界

- 无 `services/routing-config`、草稿/审批/发布 PostgreSQL、签名、分发发件箱、路由服务监视器、可信持久缓存或启用确认水位。
- 当前 v3 仍由启动配置装载；严格配置解析不是控制平面，也不是在线热发布。
- 配置与风险权限不能执行资金；没有生产发布或双人审批证据。
