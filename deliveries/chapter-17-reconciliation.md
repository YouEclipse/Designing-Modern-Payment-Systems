# 第 17 章《Reconciliation》交付记录

## 章节契约

- 核心命题：Reconciliation 追加并匹配外部证据，提出受控 Resolution；它不能按“最后到达”覆盖资金终态，也不能直接发起新资金操作。
- 演进线：补 Query → 识别来源/范围盲区 → Run/Checkpoint/Evidence/Difference → 版本化修复命令与 review。
- 本章代码基线：Gateway 已有 append-only Provider observation ledger，Payment 已有单调归约；独立周期对账 workflow 尚未实现。
- 后续范围：经济行与费用归因留给第 18 章。

## Status

- `implemented`：固定实现中的追加式 Provider observation、不可变 connection registry 与 Payment 单调冲突归约。
- `teaching_model`：正文中的 Run/Checkpoint、Difference 分类器与 Resolution proposal Go 模型。
- `target_contract`：独立周期 inventory、持久 Checkpoint/Difference/Resolution、repair receipt、人工队列与真实 Provider report adapter。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库链接需授权；不修改可见性。
- 本章没有宣称存在 `services/reconciliation`；当前 runtime 增量仅是可复用的事实账本与领域边界。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | sale/query/webhook 观察必须追加且与精确身份绑定 | `services/provider-gateway/internal/gateway/observations.go`；迁移 `0004_provider_operation_observations.sql` | `TestProviderObservationLedgerCoversSaleQueryAndWebhook`；`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable` |
| `implemented` | 冲突观察保留事实并进入 review，不 last-write-wins | Gateway reducer；Payment `domain/model.go` | `TestPostgresWebhookObservationIsAtomicWithReducerAndRetainsConflictingDelivery`；`TestConflictingTerminalOutcomeRequiresReview` |
| `implemented` | 普通应用/迁移路径下，Observation 表和 connection revision registry 的修改、删除或清空由 trigger 拒绝 | 迁移 `0003_recovery_invariants.sql`、`0004_provider_operation_observations.sql`、`0005_append_only_truncate_guards.sql` | `TestPostgresConnectionRegistryRejectsMutationOfBoundRevision`；`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`；`TestPostgresConnectionRevisionRegistryRejectsDeleteAndTruncate` |
| `teaching_model` | 分类器只提出动作、不修改 Payment | 本章 `reconciliation` 教学包 | `scripts/go-examples/tests/chapter17` compile/race/vet；classifier mutation and checkpoint edge tests |

## 图表与书稿验证

- Mermaid 覆盖来源摄取、checkpoint crash window、Payment repair command 与 review 出口。
- 本章 Go 示例以 chapter17 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；持久工作流和权限边界由时序/流程图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/reconciliation.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/reconciliation.md` 与 `git diff --check`；这些门禁只证明分类器和 checkpoint 教学合同。
- 当前 runtime 主张只能引用 manifest 中列出的 observation-ledger 代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得据此宣称周期 Reconciliation 服务已实现。

## 未实现与能力边界

- 无独立 Run、范围水位、双向 inventory、Checkpoint、Difference/Resolution schema、repair receipt 或人工运营队列。
- Gateway observation ledger 记录在线协议事实，不证明 Provider report 覆盖完整，也不是 Settlement 或 Ledger。
- 可 DROP 上述 trigger 的数据库 owner 仍是信任边界；这些门禁不是对特权管理员的防篡改证明。
- 当前只覆盖 automatic-capture SALE/Query/Simulator，不是 PSP live 对账能力。
