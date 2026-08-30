# 第 1 章《一笔支付是如何发生的》交付记录

## 章节契约

- 核心命题：一笔支付是一组可关联、不可被迟到消息抹掉的资金事实，不是一次 HTTP 调用的成功布尔值。
- 演进线：同步 `SALE` → 履约与认证异步 → 响应丢失和补偿 → 支付意图、支付尝试、资金操作与证据。
- 新增概念：授权、请款、清算、结算的阶段差异；业务投影与资金事实分离；`UNKNOWN` 和完整性复核。
- 后续范围：第 3 章精确执行身份、第 10～12 章恢复与幂等、第 17～18 章事实和经济闭环。本章不开放手动请款、撤销或退款接口。

## 能力状态

- `implemented`：固定实现中的自动请款 `SALE` 状态、持久幂等、收件箱与发件箱、`UNKNOWN` 原操作收敛和冲突完整性复核。
- `teaching_model`：正文 3 个 Go 代码块中的同步反例、最小 `SALE` 合同与追加式投影模型。
- `target_contract`：手动请款、撤销、退款、完整异步认证、账本、结算与争议流程。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 切片：令牌化银行卡自动请款 `SALE`；真实 PostgreSQL 保存三层状态、幂等、收件箱与发件箱；模拟器模拟外部支付服务商，不处理真钱。

## 主张—实现—测试

下列工程路径均相对于以上固定实现；正文示例是便于阅读的最小模型，不冒充完整存储层。

| 状态 | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | 授权不等于已请款成功 | `services/payment/internal/domain/model.go`；`services/provider-gateway/internal/gateway/service.go` | `TestSuccessfulAutomaticCaptureTransition`；`TestAuthorizedResponseDoesNotBecomeCapturedSuccess` |
| `implemented` | `UNKNOWN` 阻止第二次确认，沿原资金操作收敛 | 支付领域模型、应用对账器、服务商网关服务 | `TestUnknownBlocksNewConfirmationAndConverges`；Compose `TestPaymentVerticalSlice` |
| `implemented` | 成功之后的冲突不能覆盖资金事实 | 支付领域模型、PostgreSQL 存储层 | `TestConflictingTerminalOutcomeRequiresReview`；`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess` |
| `implemented` | 同事件编号的不同内容不是正常重复 | 支付收件箱、网关回调通知存储 | `TestDuplicateInboxEventRejectsChangedPayloadWithSameCorrelation`；正文 `TestProjectionRetainsSuccessAndFlagsConflicts` |
| `implemented` | 同步状态与幂等结果必须一起提交 | `services/payment/internal/postgres/store.go` | `TestApplyOutcomeCompletesConfirmationIdempotencyAtomically` |
| `implemented` | 生产不能开启模拟资金通道 | 网关配置与服务商模拟器 | `TestProductionRejectsSimulatorProvider`；环境标准化测试 |

## 书稿验证

- 3 个 Go 代码块全部由[可重复验证脚本](../scripts/go-examples/README.md)从当前 Markdown 提取，编译、竞态检测和静态检查通过；测试证明同步反例会误判超时。
- 4 个 Mermaid 用 `scripts/verify-mermaid.sh` 实际渲染通过。
- 1 张 Ian 风格图：`assets/payment-lifecycle-illustrations/01-success-is-not-one-fact.png`，1672×941，已经检查人物、标注和留白。
- 官方资料在正文就近引用；本地链接与资源存在性、`git diff --check` 已检查。
- 实现提交的单元测试、竞态检测、静态检查、OpenAPI、真实 PostgreSQL 集成测试、四镜像构建和 Compose 端到端测试全部通过；代码远端提交值与上述提交一致。
- 书稿版本取包含本记录的 Git 提交，不在文件中自引用尚未生成的书稿提交值。

## 能力边界

这是可运行的支付核心与外部模拟器，不是 PSP 沙箱或真实生产收款系统。示意状态图包含未来手动请款与退款投影；现有枚举、不可达预留状态与数据库基数以附录 A 和固定代码为准。账本、结算、争议与真实 PSP 凭证流程不在此切片内。
