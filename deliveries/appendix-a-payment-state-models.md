# 附录 A《支付意图、支付尝试与服务商请求状态模型》交付记录

## 合同与状态

- 核心命题：支付意图业务结果、完整性状态、支付尝试与服务商操作是不同轴；`UNKNOWN` 保留原资金身份，迟到证据只能追加归约。
- `implemented`：当前自动请款 `SALE` 的支付意图、支付尝试与操作、原操作恢复、冲突复核与精确连接身份。
- `teaching_model`：附录中的 `ValidateCapturedSale` 不变量检查器及其边界测试。
- `target_contract`：手动请款、撤销、退款、金额预留和独立后续操作；当前运行时不存在。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为私有；固定实现与绿色 CI 链接已绑定同一 SHA，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| 状态 | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | 超时或断连只把原操作置为 `UNKNOWN/possible`，不能授权新资金路径 | `services/payment/internal/domain/model.go`；`internal/recovery/planner.go`；迁移 `005_payment_recovery.sql` | `TestUnknownBlocksNewConfirmationAndConverges`；`TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart` |
| `implemented` | 已确认请款不因迟到弱证据消失；真正不相容的终态保留业务结果并进入完整性复核 | 支付领域归约器；网关仅追加观察 | `TestConflictingTerminalOutcomeRequiresReview`；`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess`；观察记录 PostgreSQL 测试 |
| `implemented` | 服务商操作固定连接、修订版、命名空间与引用，恢复不得漂移 | 支付迁移 `003_provider_connection_identity.sql`；网关修订版注册表 | `TestProviderOutcomeCannotChangePinnedRevisionOrNamespace`；`TestConnectionRegistryNewRevisionCannotRecoverOldUnknown` |
| `teaching_model` | 已请款 `SALE` 必须同时满足对象关联、操作种类、确定性、引用与金额不变量 | 附录 `package statemodel` 两个 Go 代码块 | `scripts/go-examples/tests/appendix_a/invariant_edges_test.go`，由验证器执行编译、竞态与静态检查 |
| `target_contract` | 请款、撤销和退款使用独立键与操作，绑定原连接和引用，并原子预留额度 | 附录 A.7 合同 | 当前没有实现或测试证据，不得写成已支持 API |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/payment-state-models.md`、`scripts/verify-mermaid.sh appendices/payment-state-models.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 主张才可随书发布；教学测试不能替代 PostgreSQL 与 Compose 证据。

## 能力边界

- 当前只覆盖令牌化银行卡自动请款 `SALE`，模拟器不处理真实资金；没有手动请款、撤销、退款、账本、结算或生产服务商认证。
- 状态图是平台聚合与恢复边界，不是所有卡组织、钱包或银行转账协议的统一状态机。
- 合法的 `refunded_minor` 教学分支只验证金额关系，不证明当前 API 可以退款。
