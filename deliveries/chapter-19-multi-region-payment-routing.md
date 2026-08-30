# 第 19 章《多区域支付路由》交付记录

## 章节契约

- 核心命题：多个区域可以复制入口与计算，不能复制同一支付意图的扣款权；租约帮助交接，归属世代才能拒绝旧写入者。
- 演进线：地理 DNS 与异步复制 → 单支付意图归属者 → 数据分级一致性 → 隔离、仅恢复模式与故障转移操作手册。
- 本章代码基线：已有不可变连接身份、派发与恢复证据；跨区目录、归属与隔离机制尚未实现。
- 后续范围：配置制品跨区发布与综合容量分别留给第 20、22 章。

## 能力状态

- `implemented`：固定实现中的单区域精确连接身份、派发与恢复证据，以及单 PostgreSQL 商户键获胜者。
- `teaching_model`：正文中的归属世代、受隔离命令与租约绑定接管 Go 模型。
- `target_contract`：全局支付意图目录、归属与执行区域、跨区网关权限、故障转移与回切、数据驻留和真实灾备证据。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库链接需授权；不改变可见性。
- 本章是设计型交付，没有把单库并发或区域字段冒充多区域运行时。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 已有操作不能更换连接、修订版或命名空间 | 支付迁移 `003_provider_connection_identity.sql`；网关连接注册表 | `TestProviderOutcomeCannotChangePinnedRevisionOrNamespace`；`TestConnectionRegistryNewRevisionCannotRecoverOldUnknown` |
| `implemented` | 当前运行时持久化外发与原操作恢复证据；尚无接管消费者或归属语义 | 网关迁移 `0003_recovery_invariants.sql`；支付迁移 `005_payment_recovery.sql` | `TestPostgresGatewayOperationRecoveryInvariants`；`TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart` |
| `teaching_model` | 租约与世代许可必须在写事务内验证 | 本章 `fencing` 教学包 | `scripts/go-examples/tests/chapter19` 编译、竞态与静态检查；过期归属者和租约边界测试 |
| `implemented` | 商户幂等当前只在单一 PostgreSQL 权威范围内选出唯一获胜者 | 支付服务本地幂等模式 | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath`；该证据不作全球仲裁声明 |

## 图表与书稿验证

- Mermaid 覆盖归属者写权限、接管与回切、区域数据最小化和配置自治。
- 本章 Go 示例以第 19 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；故障时序和权限关系需要精确图。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/multi-region.md`、`scripts/verify-mermaid.sh part-5-global-routing/multi-region.md` 与 `git diff --check`；这些门禁只证明隔离教学模型。
- 当前运行时主张只能引用交付清单中列出的单区域身份与恢复代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得把区域字段或单库并发冒充多区域运行证据。

## 未实现与能力边界

- 无全局支付意图目录、归属与执行区域状态、归属世代、网关跨区许可、跨区故障转移与回切，或数据驻留部署证据。
- 单 PostgreSQL、两个进程或 `region` 字段不能证明异地容灾。
- 当前系统没有生产 PSP、多区域容量或灾备 RPO/RTO 证明。
