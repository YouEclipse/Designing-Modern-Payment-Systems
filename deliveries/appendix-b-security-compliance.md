# 附录 B《安全、合规与数据边界》交付记录

## 合同与 status

- 核心命题：安全边界必须落实为最小数据、显式身份、严格输入、可轮换密钥与可失败测试；Tokenization 不自动消除 PCI DSS scope。
- `implemented`：当前 local/test 切片的严格 JSON/体积限制、内部 token、Simulator 环境门禁、精确 connection 关联、HMAC webhook、durable inbox/quarantine 与不可变 observation/revision 记录。
- `teaching_model`：附录 allowlist/audit-shape Go 检查器，只证明结构边界。
- `target_contract`：mTLS、独立 workload principal/DB role、KMS/Secret Manager、全链路 redaction、保留删除执行器、真实 PSP replay contract 与合规评估。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为 private；固定实现与绿色 CI 链接已完成同一 SHA pin，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| status | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | 未知/超大 JSON 和未授权内部请求 fail closed；非开发环境缺少 token 不启动 | Router/Payment/Gateway strict HTTP readers；`services/router/cmd/server/main.go` | strict JSON/OpenAPI、internal-token 与 non-development startup tests |
| `implemented` | webhook 验签、事件去重和 operation 关联是独立门；冲突不能覆盖资金事实 | Gateway webhook/HMAC、inbox/quarantine、observations；迁移 `0004`/`0005` | webhook auth/association、atomic inbox/reducer 与 append-only PostgreSQL tests |
| `implemented` | connection revision 是不可变执行身份，Simulator 只允许开发/测试配置 | Gateway connection registry 与 startup config | revision drift/restart tests；production-simulator fail-closed tests |
| `teaching_model` | 输出字段使用递归 allowlist、大小限制与禁名检查，且不修改调用者输入 | 附录 `package boundary` 两个 Go block | `scripts/go-examples/tests/appendix_b/boundary_edges_test.go`，由 verifier 执行 compile/race/vet |
| `target_contract` | mTLS/principal、独立生产角色、secret 生命周期、全输出面 redaction 与数据驻留必须由部署证据证明 | 附录 B.5～B.8 合同 | 当前无目标环境证据，不得从共享开发 token 或 Compose 推导完成 |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/security-compliance.md`、`scripts/verify-mermaid.sh appendices/security-compliance.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 主张才可随书发布；结构检查器不证明字段值来源、日志 redaction 或 PCI 合规。

## 能力边界

- 当前共享 token 与 HMAC 是 local/test 工程护栏，不是生产零信任、密钥托管或合规认证。
- PAN、CVV/SAD 与真实 secret 不进入测试 fixture；泄漏测试必须使用合成标记。
- Simulator、私有仓库和绿色 CI 都不能作为真实 PSP sandbox/live certification 或 PCI DSS assessment 的替代证据。
