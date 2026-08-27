# 附录 C《容量规划与上线检查清单》交付记录

## 合同与 status

- 核心命题：上线结论必须绑定版本、环境、工作负载、故障条件、资源/外部配额与资金不变量；缺少必需证据就是 No-Go。
- `implemented`：当前代码 CI 的 unit/race/vet、契约/构建、真实 PostgreSQL integration、镜像、Compose readiness 与窄支付 E2E。
- `teaching_model`：附录 `GlobalLiveGate` 只聚合已给定证据并 fail closed，不采集负载或批准上线。
- `target_contract`：统一 `test/system`、`tools/capacity-plan`、`tools/launchcheck`、目标环境容量/灾备报告、Provider live 授权与业务签核。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为 private；固定实现与绿色 CI 链接已完成同一 SHA pin，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| status | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | 当前窄切片以真实 PostgreSQL 检查迁移、事务、幂等、UNKNOWN 恢复和追加证据 | `.github/workflows/ci.yml`；Payment/Gateway integration suites | 固定 CI 的 PostgreSQL jobs；只有 pin 后可作为当前证据 |
| `implemented` | Compose E2E 覆盖 approved/declined、Risk/no-candidate 零资金、response loss 恢复与并发唯一资金路径 | `test/e2e/payment_flow_test.go`；Compose stack | `TestPaymentVerticalSlice`；`TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` 等固定 CI 证据 |
| `teaching_model` | 上线门禁要求每一项证据显式 pass、scope/version/time 一致，保留全部 failure，缺项不默认绿 | 附录 `package launchgate` 两个 Go block | `scripts/go-examples/tests/appendix_c/gate_edges_test.go`，由 verifier 执行 compile/race/vet |
| `target_contract` | 容量、外部配额、故障余量、积压追平、备份恢复、安全运营和 Go/No-Go 必须来自目标环境 | 附录 C.1～C.9 合同；未来 tools/system harness | 当前无 100k/global/live 或目标环境容量报告，不得由本机/Simulator CI 推导 |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/capacity-planning-and-launch-checklist.md`、`scripts/verify-mermaid.sh appendices/capacity-planning-and-launch-checklist.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 基线才可随书发布；该 CI 仍不是目标环境容量或 live launch certificate。

## 能力边界

- 当前没有完整 `test/system`、`tools/capacity-plan`、`tools/launchcheck`、灾备演练、真实 PSP 配额/凭证或生产安全运营签核。
- 书中算例、Simulator 和 `GlobalLiveGate` 都不能发起真实流量、测量容量或替负责人批准上线。
- 新 Provider、账户、Region、配置或 workload 会使旧证据失效；manifest 与门禁必须重新绑定新的 scope/version/time。
