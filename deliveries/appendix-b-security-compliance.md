# 附录 B《安全、合规与数据边界》交付记录

## 合同与状态

- 核心命题：安全边界必须落实为最小数据、显式身份、严格输入、可轮换密钥与可失败测试；令牌化不自动消除 PCI DSS 范围。
- `implemented`：当前本地与测试切片的严格 JSON 与体积限制、内部令牌、模拟器环境门禁、精确连接关联、HMAC 回调、持久收件箱与隔离区，以及不可变观察与修订版记录。
- `teaching_model`：附录允许列表与审计形状 Go 检查器，只证明结构边界。
- `target_contract`：mTLS、独立工作负载主体与数据库角色、KMS 与秘密管理器、全链路脱敏、保留删除执行器、真实 PSP 重放合同与合规评估。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为私有；固定实现与绿色 CI 链接已绑定同一 SHA，但仍需获授权的 GitHub 身份访问。

## 主张—代码/示例—测试映射

| 状态 | 主张或不变量 | 代码或教学示例 | 测试/验证证据 |
| --- | --- | --- | --- |
| `implemented` | 未知或超大 JSON 和未授权内部请求缺证据即拒绝；非开发环境缺少令牌不启动 | 路由、支付与网关服务的严格 HTTP 读取器；`services/router/cmd/server/main.go` | 严格 JSON 与 OpenAPI、内部令牌和非开发环境启动测试 |
| `implemented` | 回调验签、事件去重和操作关联是独立门禁；冲突不能覆盖资金事实 | 网关回调与 HMAC、收件箱与隔离区、观察记录；迁移 `0004`/`0005` | 回调认证与关联、原子收件箱与归约器、仅追加 PostgreSQL 测试 |
| `implemented` | 连接修订版是不可变执行身份，模拟器只允许开发与测试配置 | 网关连接注册表与启动配置 | 修订版漂移与重启测试；生产模拟器缺证据即拒绝测试 |
| `teaching_model` | 输出字段使用递归允许列表、大小限制与禁名检查，且不修改调用者输入 | 附录 `package boundary` 两个 Go 代码块 | `scripts/go-examples/tests/appendix_b/boundary_edges_test.go`，由验证器执行编译、竞态与静态检查 |
| `target_contract` | mTLS 与主体、独立生产角色、秘密生命周期、全输出面脱敏与数据驻留必须由部署证据证明 | 附录 B.5～B.8 合同 | 当前无目标环境证据，不得从共享开发令牌或 Compose 推导完成 |

## 发布验证

- 书稿侧：`go run scripts/verify-go-examples.go appendices/security-compliance.md`、`scripts/verify-mermaid.sh appendices/security-compliance.md`、本地链接检查与 `git diff --check`。
- 代码侧：只有固定 `bc4afde24da708d22f67da72a90a1557b1477173` 对应的 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 为绿色后，表中 `implemented` 主张才可随书发布；结构检查器不证明字段值来源、日志脱敏或 PCI 合规。

## 能力边界

- 当前共享令牌与 HMAC 是本地和测试工程护栏，不是生产零信任、密钥托管或合规认证。
- PAN、CVV/SAD 与真实秘密不进入测试夹具；泄漏测试必须使用合成标记。
- 模拟器、私有仓库和绿色 CI 都不能作为真实 PSP 沙箱与生产认证或 PCI DSS 评估的替代证据。
