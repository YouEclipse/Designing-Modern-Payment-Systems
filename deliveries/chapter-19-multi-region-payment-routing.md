# 第 19 章《Multi-Region Payment Routing》交付记录

## 章节契约

- 核心命题：多 Region 可以复制入口与计算，不能复制同一 Intent 的扣款权；lease 帮助交接，ownership epoch 才能拒绝旧 writer。
- 演进线：Geo DNS + 异步复制 → 单 Intent home owner → 数据分级一致性 → fencing、recovery-only 与 failover Runbook。
- 本章代码基线：已有 immutable connection identity、dispatch/recovery evidence；跨区目录、ownership 与 fencing 尚未实现。
- 后续范围：配置制品跨区发布与综合容量分别留给第 20、22 章。

## Status

- `implemented`：固定实现中的单 Region 精确 connection identity、dispatch/recovery 证据与单 PostgreSQL merchant-key winner。
- `teaching_model`：正文中的 ownership epoch、fenced command 与 lease-bound takeover Go 模型。
- `target_contract`：Global Intent Directory、Home/Execution Region、跨区 Gateway 权限、failover/failback、数据驻留和真实灾备证据。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库链接需授权；不改变可见性。
- 本章是设计型交付，没有把单库并发或 Region 字段冒充多 Region runtime。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 已有 Operation 不能更换 connection/revision/namespace | Payment migration `003_provider_connection_identity.sql`；Gateway connection registry | `TestProviderOutcomeCannotChangePinnedRevisionOrNamespace`；`TestConnectionRegistryNewRevisionCannotRecoverOldUnknown` |
| `implemented` | 当前 runtime 持久化外发与原操作恢复证据；尚无 takeover consumer 或 ownership 语义 | Gateway migration `0003_recovery_invariants.sql`；Payment migration `005_payment_recovery.sql` | `TestPostgresGatewayOperationRecoveryInvariants`；`TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart` |
| `teaching_model` | lease/epoch 许可必须在写事务内验证 | 本章 `fencing` 教学包 | `scripts/go-examples/tests/chapter19` compile/race/vet；stale owner and lease boundary tests |
| `implemented` | merchant idempotency 当前只在单 PostgreSQL authority 内选出唯一 winner | Payment 本地 idempotency schema | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath`；该证据不作全球仲裁声明 |

## 图表与书稿验证

- Mermaid 覆盖 owner 写权限、takeover/failback、区域数据最小化与配置自治。
- 本章 Go 示例以 chapter19 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；故障时序和 authority 关系需要精确图。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/multi-region.md`、`scripts/verify-mermaid.sh part-5-global-routing/multi-region.md` 与 `git diff --check`；这些门禁只证明 fencing 教学模型。
- 当前 runtime 主张只能引用 manifest 中列出的单 Region identity/recovery 代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得把 Region 字段或单库并发冒充多 Region 运行证据。

## 未实现与能力边界

- 无 Global Intent Directory、Home/Execution Region 状态、ownership epoch、Gateway 跨区许可、跨区 failover/failback 或 residency deployment evidence。
- 单 PostgreSQL、两个进程或 `region` 字段不能证明异地容灾。
- 当前系统没有 live PSP、多 Region 容量或灾备 RPO/RTO 证明。
