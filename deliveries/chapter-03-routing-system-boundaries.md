# 第 3 章《支付路由系统的职责与边界》交付记录

## 章节契约

- 核心命题：路由服务决定精确路径，支付服务拥有状态与资金操作，服务商网关只执行被固定的连接；三个角色不能暗中互相代替。
- 演进线：支付服务商字符串 → 同一 PSP 的多账户与多区域 → 决策与结果分离 → 不可变执行身份和恢复边界。
- 本章增量：连接编号、修订和命名空间贯穿路由服务 → 支付服务 → 服务商网关 → PostgreSQL；注册表拒绝旧修订重绑定；别名只提供显式的旧操作恢复。
- 后续范围：路由上下文、规则集、能力、评分、后备路由和决策存储不属于本章的完成声明；它们的当前状态分别以第 4～14 章交付记录为准。

## 能力状态

- `implemented`：固定实现中的精确连接、修订与命名空间固定，不可变注册表、显式恢复别名与 `UNKNOWN` 单路径。
- `teaching_model`：正文 3 个 Go 代码块中的反例，以及路由服务、支付服务与服务商网关的合同和执行身份固定模型。
- `target_contract`：真实适配器的账户、区域、接口和凭证作用域，跨支付服务商后备执行器与独立决策存储；路由上下文和评分的当前状态以对应后章交付记录为准。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 接口：路由服务 v1 决策输出完整的已选候选通道；服务商网关的执行、查询和结果接口校验同一个身份元组。
- 迁移：支付服务 `003` 与服务商网关 `0002` 是协调契约升级，不能与仍写旧身份的进程混跑；旧版恢复记录与请求载荷哈希升级均有集成测试。

## 主张—实现—测试

工程文件相对于固定实现。

| 状态 | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | 决策携带可执行连接，执行层不二次选路 | 路由类型与引擎；支付应用与客户端；网关服务 | 路由服务单元测试；`TestSameProviderConnectionsStayPinnedThroughUnknownRecovery` |
| `implemented` | 已绑定的修订不能更换端点或命名空间 | 服务商网关 `connection_registry.go` 与 PostgreSQL 注册表 | `TestPostgresConnectionRegistryRejectsMutationOfBoundRevision`；批次回滚与并发胜出项测试 |
| `implemented` | 新修订不能恢复旧 `UNKNOWN` 操作 | 网关就绪检查、注册表与服务 | `TestPostgresConnectionRegistryNewRevisionCannotRecoverOldUnknown`；`TestReadinessRejectsRevisionAndNamespaceDrift` |
| `implemented` | 相同命名空间是冲突域，不是回调通知通行证 | 网关恢复别名与回调通知验证 | `TestRecoveryAliasCanonicalizesWebhookTupleAndSharesNamespaceCollisionDomain`；跨连接回调通知拒绝测试 |
| `implemented` | 支付尝试与资金操作必须固定一致身份 | 支付领域模型、数据库复合外键 | `TestPrepareConfirmationRequiresMatchingExecutionIdentity`；`TestDatabaseRejectsOperationExecutionIdentityDrift` |
| `implemented` | 同一服务商引用跨别名碰撞需要隔离 | 支付收件箱与引用唯一性约束 | `TestProviderReferenceCollisionSpansAliasesInTheSameNamespace`；不同命名空间可重复的对照测试 |
| `implemented` | 旧版升级不能把历史操作变成新路径 | 支付服务与网关迁移、哈希升级 | `TestProviderConnectionMigrationUpgradesLegacyOutcomePayloads`；`TestMemoryStoreUpgradesExactLegacyRequestHash` |
| `implemented` | `UNKNOWN` 不增加资金路径且不改变连接 | `test/e2e/payment_flow_test.go` | Compose：一个支付尝试、一项支付资金操作和一项网关资金操作；两端的编号、修订与命名空间相等 |

## 书稿验证

- 3 个 Go 代码块全部提取为独立包，编译、竞态检测和静态检查通过；`TestPinRouteBindsExactExecutionIdentity` 检查元组固定和缺字段拒绝。
- 5 个 Mermaid 已实际渲染；包括边界、基数、时序与控制平面关系。
- 1 张 Ian 配图：`assets/routing-system-boundaries-illustrations/01-decision-is-not-execution.png`，1672×941，已视觉检查。
- PAN 与 CHD、CVV 与 SAD 的分类已校正；服务商幂等、原引用绑定、令牌作用域由官方资料核验并就近引用。
- 本地链接、资源路径和差异格式检查通过；固定实现的完整单元测试、竞态检测、静态检查、契约检查、集成测试与 Compose 门禁全部通过。
- 代码远端提交值已核对；书稿版本由包含本记录的 Git 历史确定。

## 能力边界

默认两个候选通道都由开发模拟器执行，共享明确的模拟服务商命名空间，不代表两个真实 PSP 账户。注册表当前使用指纹保护模拟适配器的端点、服务商、命名空间和验签配置；真实适配器还必须纳入真实账户、区域、接口版本与凭证作用域，不能以本地指纹代替支付服务商契约。同一固定提交已经包含后章的版本化路由上下文与 v3 评分，但它们不扩大本章的完成声明；系统仍没有真钱接入、手动请款、撤销、退款或跨支付服务商后备执行器。
