# 附录 C《容量规划与上线检查清单》交付记录

## 合同与状态

- 核心命题：上线结论必须绑定版本、环境、工作负载、故障条件、资源与外部配额及资金不变量；缺少必需证据就不发布。
- `implemented`：当前代码 CI 的单元测试、竞态检测、静态检查、契约与构建、真实 PostgreSQL 集成、镜像、Compose 就绪检查与窄支付端到端测试。
- `teaching_model`：附录 `GlobalLiveGate` 只聚合已给定证据并缺证据即拒绝，不采集负载或批准上线。
- `target_contract`：统一 `test/system`、`tools/capacity-plan`、`tools/launchcheck`、目标环境容量与灾备报告、服务商生产授权与业务签核。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为私有；固定实现与绿色 CI 链接已绑定同一 SHA，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| 状态 | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | 当前窄切片以真实 PostgreSQL 检查迁移、事务、幂等、`UNKNOWN` 恢复和追加证据 | `.github/workflows/ci.yml`；支付与网关集成测试套件 | 固定 CI 的 PostgreSQL 任务；只有绑定固定版本后才可作为当前证据 |
| `implemented` | Compose 端到端测试覆盖批准、拒绝、风险与无候选零资金路径、响应丢失恢复和并发唯一资金路径 | `test/e2e/payment_flow_test.go`；Compose 服务栈 | `TestPaymentVerticalSlice`；`TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` 等固定 CI 证据 |
| `teaching_model` | 上线门禁要求每一项证据显式通过，范围、版本和时间一致，保留全部失败，缺项不默认通过 | 附录 `package launchgate` 两个 Go 代码块 | `scripts/go-examples/tests/appendix_c/gate_edges_test.go`，由验证器执行编译、竞态与静态检查 |
| `target_contract` | 容量、外部配额、故障余量、积压追平、备份恢复、安全运营和上线决策必须来自目标环境 | 附录 C.1～C.9 合同；未来系统工具编排器 | 当前无 10 万、全球、生产或目标环境容量报告，不得由本机或模拟器 CI 推导 |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/capacity-planning-and-launch-checklist.md`、`scripts/verify-mermaid.sh appendices/capacity-planning-and-launch-checklist.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 基线才可随书发布；该 CI 仍不是目标环境容量或生产上线证书。

## 能力边界

- 当前没有完整 `test/system`、`tools/capacity-plan`、`tools/launchcheck`、灾备演练、真实 PSP 配额与凭证或生产安全运营签核。
- 书中算例、模拟器和 `GlobalLiveGate` 都不能发起真实流量、测量容量或替负责人批准上线。
- 新服务商、账户、区域、配置或工作负载会使旧证据失效；交付清单与门禁必须重新绑定新的范围、版本和时间。
