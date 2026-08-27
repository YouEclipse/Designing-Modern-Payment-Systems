# 附录 A《Payment Intent、Attempt 与 Provider Request 状态模型》交付记录

## 合同与 status

- 核心命题：Intent 业务结果、完整性状态、Attempt 与 Provider Operation 是不同轴；`UNKNOWN` 保留原资金身份，迟到证据只能追加归约。
- `implemented`：当前 automatic-capture `SALE` 的 Intent/Attempt/Operation、原 Operation 恢复、冲突 review 与精确 connection identity。
- `teaching_model`：附录中的 `ValidateCapturedSale` 不变量检查器及其边界测试。
- `target_contract`：manual Capture/Void/Refund、金额预留和独立后续 Operation；当前 runtime 不存在。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为 private；固定实现与绿色 CI 链接已完成同一 SHA pin，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| status | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | timeout/断连只把原 Operation 置为 `UNKNOWN/possible`，不能授权新资金路径 | `services/payment/internal/domain/model.go`；`internal/recovery/planner.go`；迁移 `005_payment_recovery.sql` | `TestUnknownBlocksNewConfirmationAndConverges`；`TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart` |
| `implemented` | 已确认 capture 不因迟到弱证据消失；真正不相容的终态保留业务结果并进入 integrity review | Payment domain reducer；Gateway append-only observations | `TestConflictingTerminalOutcomeRequiresReview`；`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess`；observation PostgreSQL tests |
| `implemented` | Provider Operation 固定 connection/revision/namespace/reference，恢复不得漂移 | Payment migration `003_provider_connection_identity.sql`；Gateway revision registry | `TestProviderOutcomeCannotChangePinnedRevisionOrNamespace`；`TestConnectionRegistryNewRevisionCannotRecoverOldUnknown` |
| `teaching_model` | captured SALE 必须同时满足对象关联、operation kind、certainty、reference 与金额不变量 | 附录 `package statemodel` 两个 Go block | `scripts/go-examples/tests/appendix_a/invariant_edges_test.go`，由 verifier 执行 compile/race/vet |
| `target_contract` | Capture/Void/Refund 使用独立 key/Operation，绑定原 connection/reference，并原子预留额度 | 附录 A.7 合同 | 当前没有实现或测试证据，不得写成已支持 API |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/payment-state-models.md`、`scripts/verify-mermaid.sh appendices/payment-state-models.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 主张才可随书发布；教学测试不能替代 PostgreSQL/Compose 证据。

## 能力边界

- 当前只覆盖 tokenized card automatic-capture `SALE`，Simulator 不处理真实资金；没有 manual Capture/Void/Refund、Ledger、Settlement 或 live Provider certification。
- 状态图是平台聚合与恢复边界，不是所有卡组织、钱包或银行转账协议的统一状态机。
- 合法的 `refunded_minor` 教学分支只验证金额关系，不证明当前 API 可以退款。
