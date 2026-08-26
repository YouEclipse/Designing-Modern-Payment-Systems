# 第 3 章《Payment Routing System 的职责与边界》交付记录

## 章节契约

- 核心命题：Router 决定精确路径，Payment 拥有状态与资金操作，Gateway 只执行被固定的连接；三个角色不能暗中互相代替。
- 演进线：Provider 字符串 → 同 PSP 多账户/Region → 决策与结果分离 → 不可变执行身份和恢复边界。
- 本章增量：connection ID、revision、namespace 贯穿 Router → Payment → Gateway → PostgreSQL；注册表拒绝旧修订重绑定；别名只提供显式旧操作恢复。
- 后续范围：Context、RuleSet、Capability 的完整版本化交给第 4～6 章；Scoring、Fallback 和 Decision Store 不在此提交内。

## 配套实现

- 固定实现：[adbdf75589dcb7b9ae088aa9f92fde2c2f063aeb](https://github.com/YouEclipse/payments/commit/adbdf75589dcb7b9ae088aa9f92fde2c2f063aeb)。
- CI：[32919381068 — success](https://github.com/YouEclipse/payments/actions/runs/32919381068)。
- API：Router v1 Decision 输出完整 Selected Candidate；Gateway execute/query/outcome 校验同一个身份元组。
- 迁移：Payment `003` 与 Gateway `0002` 是协调契约升级，不能与仍写旧身份的进程混跑；legacy recovery 与 payload hash 升级有集成测试。

## 主张—实现—测试

工程文件相对于固定实现。

| 主张 | 工程文件 | 验证 |
| --- | --- | --- |
| 决策携带可执行连接，执行层不二次选路 | Router routing types/engine；Payment app/client；Gateway service | Router unit；`TestSameProviderConnectionsStayPinnedThroughUnknownRecovery` |
| 已绑定的 revision 不能换 endpoint/namespace | Gateway `connection_registry.go` 与 PostgreSQL registry | `TestPostgresConnectionRegistryRejectsMutationOfBoundRevision`；batch rollback / concurrent winner tests |
| 新修订不能恢复旧 UNKNOWN | Gateway readiness / registry / service | `TestPostgresConnectionRegistryNewRevisionCannotRecoverOldUnknown`；`TestReadinessRejectsRevisionAndNamespaceDrift` |
| 相同 namespace 是冲突域，不是 webhook 通行证 | Gateway recovery aliases、webhook verification | `TestRecoveryAliasCanonicalizesWebhookTupleAndSharesNamespaceCollisionDomain`；跨连接 webhook 拒绝测试 |
| Attempt 与 Operation 必须固定一致身份 | Payment domain、数据库 composite foreign key | `TestPrepareConfirmationRequiresMatchingExecutionIdentity`；`TestDatabaseRejectsOperationExecutionIdentityDrift` |
| 同 Provider reference 跨别名碰撞需要隔离 | Payment inbox/reference uniqueness | `TestProviderReferenceCollisionSpansAliasesInTheSameNamespace`；不同 namespace 可重复的对照测试 |
| legacy 升级不能把历史操作变成新路径 | Payment/Gateway migrations 与 hash upgrade | `TestProviderConnectionMigrationUpgradesLegacyOutcomePayloads`；`TestMemoryStoreUpgradesExactLegacyRequestHash` |
| UNKNOWN 不增加资金路径且不改变连接 | `test/e2e/payment_flow_test.go` | Compose：一个 Attempt、一个 Payment Operation、一个 Gateway Operation；两端 ID/revision/namespace 相等 |

## 书稿验证

- 3 个 Go 块全部提取为独立 package，compile/race/vet 通过；`TestPinRouteBindsExactExecutionIdentity` 检查 tuple 固定和缺字段拒绝。
- 5 个 Mermaid 已实际渲染；包括边界、基数、时序与控制平面关系。
- 1 张 Ian 配图：`assets/routing-system-boundaries-illustrations/01-decision-is-not-execution.png`，1672×941，已视觉检查。
- PAN/CHD 与 CVV/SAD 分类已校正；Provider 幂等、原 reference 绑定、Token scope 由官方资料核验并就近引用。
- 本地链接、资源路径、diff-check 通过；固定实现的完整 unit/race/vet/contracts/integration/Compose 门禁全绿。
- 代码远端 SHA 已核对；书稿 revision 由包含本记录的 Git 历史确定。

## 能力边界

默认两个 Candidate 都由开发 Simulator 执行，共享明确的模拟 Provider namespace，不代表两个真实 PSP 账户。注册表当前 fingerprint 保护模拟 Adapter 的 endpoint/provider/namespace/验签配置；真实 Adapter 还必须纳入真实账户、Region、API 版本与 credential scope，不能以本地 fingerprint 代替供应商契约。没有真钱接入、manual capture、Void、Refund、完整 Context 或跨 Provider fallback。
