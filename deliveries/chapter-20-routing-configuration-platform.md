# 第 20 章《Routing Configuration Platform》交付记录

## 章节契约

- 核心命题：Bundle、Release 与 Activation 是三种不同事实；审批固定不可变内容，运行实例一次只启用完整 snapshot。
- 演进线：环境变量/文件 → 跨制品依赖闭包 → approval state machine → 原子 activation、LKG 与有限 kill switch。
- 本章代码基线：Router v3 已严格编译完整 Cost/Performance/Scoring data-plane 配置；独立 routing-config 控制面尚未实现。
- 后续范围：Model artifact 作为一种受治理制品留给第 21 章。

## Status

- `implemented`：固定实现中的严格 v3 启动配置、完整 Optimization 依赖快照、与 caller 输入深拷贝隔离和确定重放。
- `teaching_model`：正文中的 Bundle/Release/Activation 状态机与 atomic runtime swap Go 模型。
- `target_contract`：独立 routing-config 服务、双人审批、签名、分发 outbox、可信缓存、watcher、activation ACK 与跨区发布。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库链接需授权；不改变可见性。
- 本章没有虚构 `services/routing-config`、审批数据库或 watcher runtime。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 运行请求不能提交自己的 Policy/Artifact | Router OpenAPI 0.4；`services/router/internal/httpapi` strict reader | strict JSON/OpenAPI contract tests；v3 golden fixture |
| `implemented` | v3 启动快照必须完整，Risk/Cost/Performance/Scoring 依赖共同生效 | `services/router/cmd/server/v3_config.go`；`decision/optimization.go` | `TestDevelopmentV3IsRiskGatedAndUsesExactLocalConnections`；`TestV2AndV3ConfigurationReadersCannotBeConfused` |
| `implemented` | 同一 Engine 的 Artifact snapshot 可确定重放且不受别名修改 | `services/router/internal/{cost,performance,scoring,decision}` | canonical ownership/concurrent replay tests；`TestOptimizationSnapshotDigestOrderOwnershipAndConcurrentReplay` |
| `teaching_model` | Bundle/Release/Activation 生命周期仍是教学合同 | 本章 `configrelease` / `runtimeconfig` Go 包 | `scripts/go-examples/tests/chapter20` compile/race/vet；state-machine and atomic-swap tests |

## 图表与书稿验证

- Mermaid 覆盖 ownership、审批状态机、发布/activation 分离与完整 runtime swap。
- 本章 Go 示例以 chapter20 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；发布代际与原子性由状态/时序图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/configuration-platform.md`、`scripts/verify-mermaid.sh part-5-global-routing/configuration-platform.md` 与 `git diff --check`；这些门禁只证明 release/runtime-config 教学模型。
- 当前 runtime 主张只能引用 manifest 中列出的 v3 启动快照代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；严格启动解析不等于控制平面或热发布。

## 未实现与能力边界

- 无 `services/routing-config`、Draft/Approval/Release PostgreSQL、签名、分发 outbox、Router watcher、可信持久缓存或 activation ACK 水位。
- 当前 v3 仍由启动配置装载；严格配置解析不是控制平面，也不是在线热发布。
- Config/Risk 权限不能执行资金；没有 live rollout 或双人审批证据。
