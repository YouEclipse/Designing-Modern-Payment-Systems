# 第 17 章《对账》交付记录

## 章节契约

- 核心命题：对账追加并匹配外部证据，提出受控修复方案；它不能按“最后到达”覆盖资金终态，也不能直接发起新资金操作。
- 演进线：补充查询 → 识别来源与范围盲区 → 运行、检查点、证据与差异 → 版本化修复命令与人工复核。
- 本章代码基线：网关已有仅追加的服务商观察账本，支付服务已有单调归约；独立周期对账工作流尚未实现。
- 后续范围：经济行与费用归因留给第 18 章。

## 能力状态

- `implemented`：固定实现中的追加式服务商观察、不可变连接注册表与支付服务单调冲突归约。
- `teaching_model`：正文中的运行与检查点、差异分类器和修复提案 Go 模型。
- `target_contract`：独立周期清单、持久检查点、差异与修复方案、修复回执、人工队列和真实服务商报告适配器。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库链接需授权；不修改可见性。
- 本章没有宣称存在 `services/reconciliation`；当前运行时增量仅是可复用的事实账本与领域边界。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | `sale`、查询与回调观察必须追加且与精确身份绑定 | `services/provider-gateway/internal/gateway/observations.go`；迁移 `0004_provider_operation_observations.sql` | `TestProviderObservationLedgerCoversSaleQueryAndWebhook`；`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable` |
| `implemented` | 冲突观察保留事实并进入人工复核，不采用后写覆盖前写 | 网关归约器；支付服务 `domain/model.go` | `TestPostgresWebhookObservationIsAtomicWithReducerAndRetainsConflictingDelivery`；`TestConflictingTerminalOutcomeRequiresReview` |
| `implemented` | 普通应用与迁移路径下，观察表和连接修订版注册表的修改、删除或清空由触发器拒绝 | 迁移 `0003_recovery_invariants.sql`、`0004_provider_operation_observations.sql`、`0005_append_only_truncate_guards.sql` | `TestPostgresConnectionRegistryRejectsMutationOfBoundRevision`；`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`；`TestPostgresConnectionRevisionRegistryRejectsDeleteAndTruncate` |
| `teaching_model` | 分类器只提出动作、不修改支付状态 | 本章 `reconciliation` 教学包 | `scripts/go-examples/tests/chapter17` 编译、竞态与静态检查；分类器修改保护和检查点边界测试 |

## 图表与书稿验证

- Mermaid 覆盖来源摄取、检查点崩溃窗口、支付修复命令与人工复核出口。
- 本章 Go 示例以第 17 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；持久工作流和权限边界由时序/流程图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/reconciliation.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/reconciliation.md` 与 `git diff --check`；这些门禁只证明分类器和检查点教学合同。
- 当前运行时主张只能引用交付清单中列出的观察账本代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得据此宣称周期对账服务已实现。

## 未实现与能力边界

- 无独立运行记录、范围水位、双向清单、检查点、差异与修复方案模式、修复回执或人工运营队列。
- 网关观察账本记录在线协议事实，不证明服务商报告覆盖完整，也不是结算或账本。
- 可删除上述触发器的数据库所有者仍是信任边界；这些门禁不是对特权管理员的防篡改证明。
- 当前只覆盖自动请款 `SALE`、查询与模拟器，不是 PSP 生产对账能力。
