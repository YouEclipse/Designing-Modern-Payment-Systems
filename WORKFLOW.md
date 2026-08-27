# 《Designing Modern Payment Systems》写作与实现工作流

这本书采用“双仓库、单交付”的方式推进：每个章节既要形成可独立阅读的系统设计，也要让配套 Go 支付系统向前演进一小步。正文、示例代码、真实实现、图表和测试必须描述同一套支付语义。

## 1. 两个仓库

| 内容 | 本地路径 | GitHub |
| --- | --- | --- |
| 中文书稿与插图 | `/Users/yoyo/workspace/payment-book-active` | [YouEclipse/Designing-Modern-Payment-Systems](https://github.com/YouEclipse/Designing-Modern-Payment-Systems) |
| Go 微服务实现 | `/Users/yoyo/workspace/payments-system-active` | [YouEclipse/payments](https://github.com/YouEclipse/payments) |

开始前必须分别确认 Git root、工作树、remote 和分支，保留已有修改。结束时分别提交、推送并核对本地与远端 SHA。

本项目所有新产物都放在 `~/workspace` 下。章节、图片、脚本与正式验证报告进入对应仓库；临时 module、渲染文件和隔离 worktree 使用书仓 `.validation-tmp/` 或 `~/workspace/.payment-book-tmp/`。Go/Chromium 等验证子进程的临时目录也指向这些位置，不将本书产物写回 Documents/iCloud 或系统 `/tmp`。原有 Documents 副本保留，但不再作为活动工作树。

## 2. 每章先定义交付契约

写作前先回答七个问题：

1. 本章唯一的核心命题是什么？
2. 读者从哪个足够简单的系统开始？
3. 哪些规模、约束或失败场景让简单设计失效？
4. 本章自然推导出哪个新组件或不变量？
5. 哪些内容必须留给后续章节，避免提前透支？
6. 配套 Go 系统增加哪条可运行的纵向能力？
7. 哪个失败场景可以证明设计不是纸上谈兵？

完整章节通常以 5,000～8,000 中文字为目标，具体长度服从内容。目录编号、术语和跨章引用在动笔前固定；若确需改章号，先全局检查引用。

## 3. 章节叙事结构

主要章节使用同一条演进链：

```text
Problem
→ Simple Design
→ Problems at Scale
→ Production Design
→ Failure Modes
→ Trade-offs
→ Summary
```

从真实工程案例推动架构出现。先展示硬编码、单 Provider、同步响应或简单状态为什么在规模和故障下失效，再引出规则、Eligibility、Attempt、UNKNOWN、Decision Store 或 Reconciliation。不要把章节写成术语列表，也不要先展示最终大架构再倒序解释。

所有后端示例统一使用 Go。JSON、SQL、HTTP 与 Mermaid 分别用于契约、存储、接口和交互，不发展成第二套实现语言。关键示例应尽可能可编译、可测试，并链接到配套仓库的真实实现。

## 4. 支付语义红线

每次写作和代码评审都检查：

- Authorization 不等于 Capture，Capture 不等于 Settlement。
- Capture、Void、Refund 必须绑定原 Provider connection 与 transaction reference，不重新路由。
- Timeout、断连、响应丢失和含糊 5xx 是 `UNKNOWN`，不得直接 fallback。
- 同一逻辑资金操作的网络重试必须复用原 Provider、账户、区域、幂等键和 payload。
- Provider 幂等只在其契约声明的作用域内成立，跨 PSP 没有全局去重。
- Fallback 是 Attempt 之后的恢复策略；只有旧操作明确无副作用并获规则允许时，才可创建新 Attempt。
- Eligibility 硬约束必须先于 Scoring 软偏好。
- Router 只生成可审计决策；Orchestrator 拥有支付状态并执行资金副作用。
- 历史成功率带有流量构成与路由选择偏差，不能直接解释为 Provider 的因果优势。
- Tokenization 不自动移出 PCI scope；PAN、CVV/SAD 不得进入平台存储、日志或事件。
- Merchant key、operation ID、Provider idempotency key 与 Provider transaction reference 必须分开建模。
- 所有迟到资金证据都必须追加保存，不得 last-write-wins；重复或更弱且不冲突的观察按领域归约 no-op，只有不相容、身份冲突或无法安全归属的证据才进入 `review_required`。

涉及 Provider 行为、API、标准或合规时，优先引用当前官方资料并在线核验。

## 5. 正文与 Go 实现同步推进

每章只实现一个窄而完整的纵向切片。例如：

| 章节主题 | 正文交付 | 配套实现交付 |
| --- | --- | --- |
| Router 边界 | Decision 与资金副作用分离 | Router API、Eligibility、可解释排除原因 |
| Timeout / UNKNOWN | 不确定状态与恢复顺序 | 同 operation 查询、webhook 收敛、禁止新 Attempt |
| Idempotency | 三层幂等作用域 | API record、Provider key、inbox/outbox 与唯一约束 |
| Reconciliation | 多源事实与差异分类 | durable job、差异状态、人工复核出口 |

核心状态必须用真实 PostgreSQL 持久化；Simulator 只模拟外部 PSP，且必须在 production/live 环境拒绝启动。不要用内存 mock 伪装持久化、迁移、并发或 crash recovery 已完成。

如果章节暂时没有合理的代码增量，应在交付说明中明确原因，同时运行现有系统回归测试，确保文字没有破坏既有实现语义。

## 6. 图表与 Ian 风格正文配图

图形必须解决文字难以清晰表达的关系：

- 服务边界和数据流使用 Mermaid flowchart；
- timeout、webhook、query 和恢复使用 sequence diagram；
- 状态变化使用 state diagram；
- 多字段策略比较使用表格。

Ian 风格配图用于建立直觉和章节节奏，不替代精确架构图。先形成 shot list，再生成 16:9 图片，检查人物一致性、中文标注、尺寸与留白，保存到章节专属 `assets/` 目录并用相对路径嵌入。无信息增益的装饰图不进入正文。

## 7. 发布前独立审计

至少从五个视角做交叉检查：

1. **章节架构**：核心命题、因果链、跨章边界和重复内容。
2. **支付准确性**：资金语义、状态机、不变量、失败分类和术语。
3. **实现一致性**：正文、Go 代码、OpenAPI、README 和数据库模型是否一致。
4. **可靠性**：并发、crash window、重复、乱序、UNKNOWN、manual review 和 outbox/inbox。
5. **发布准备**：本地链接、Mermaid、图片、secrets、环境门禁、CI 和能力声明。

审计发现的问题必须回到正文、代码或测试中闭环，而不是只写进最终说明。

## 8. 验证清单

代码仓库至少执行：

```bash
gofmt -w services tools test
go mod tidy
git diff --check
go mod tidy -diff
go test -count=1 ./...
go test -count=1 -race ./...
go vet ./...
make contracts-check
make build
```

继续验证 integration/e2e build tags。具备 Docker/PostgreSQL 时本地运行真实集成与 Compose E2E；本机不具备时，GitHub Actions 必须负责以下最终门禁：

- 真实 PostgreSQL 迁移与 integration tests；
- 四个服务镜像构建和 readiness；
- approved、declined、timeout→`UNKNOWN`→成功收敛；
- 商户幂等与重复 confirm；
- UNKNOWN 全程只有一个 Attempt 与一个 Provider Operation 资金路径。

CI 失败后读取精确日志，修复根因并补回归测试，重新推送直到最新 master run 全绿。

书稿仓库还要检查：

- GitBook 本地 Markdown 链接全部存在；
- Go 示例登记到 `scripts/go-examples/manifest.json`，用 `go run scripts/verify-go-examples.go <chapter.md>...` 从当前正文提取并执行 compile/race/vet；新增未登记代码块必须失败，临时目录里曾经编译过不等于可重复门禁；
- Mermaid 用 `scripts/verify-mermaid.sh <chapter.md>...` 抽取并实际渲染；不能以人工阅读或代码块计数代替。脚本在单图计时前按 `MERMAID_SETUP_TIMEOUT_SECONDS` 预热并核验固定版本 CLI，把 npm/Puppeteer 缓存与 setup/render 的 `TMPDIR` 固定到书仓 `.validation-tmp/`，并用 `MERMAID_NPM_FETCH_TIMEOUT_MS` / `MERMAID_NPM_FETCH_RETRIES` 限制冷启动网络等待；默认 npx 路径总是固定 `@mermaid-js/mermaid-cli@11.16.0`，PATH 中偶然存在的 `mmdc` 或单独设置 `MERMAID_CLI_VERSION` 都不会绕过该 pin；已有受控安装只能通过 `MMDC_BIN` 显式指定，其 `--version` 必须等于 `MERMAID_CLI_VERSION`（未设置时为 pin 版本 `11.16.0`）；
- 插图尺寸、路径和 Markdown 嵌入正确；
- 术语表、SUMMARY、README 与章节链接同步；
- 不把骨架章节、Simulator 或 PSP sandbox 夸大成完成的 live 系统。

## 9. 双仓库发布顺序

1. 完成正文、Go 实现、图表、测试和审计。
2. 在代码仓库 stage 明确路径，检查 cached diff，提交并推送 `master`。
3. 等待代码仓库最新 GitHub Actions 全绿；失败时先修复代码与测试，不发布书稿中的完成声明。
4. 使用已验证的实现 SHA 更新正文固定链接，并在 `deliveries/` 写入本章交付记录。
5. 在书稿仓库检查链接、图表与最终 diff，stage 明确路径，提交并推送 `master`。
6. 用 `git ls-remote` 核对两个远端 SHA。
7. 最终交付说明包含双仓 commit、CI 链接、完成范围、未完成章节和 live-payment 边界。

每个 `deliveries/<chapter-slug>.md` 至少记录：章节核心命题、固定的实现 commit、CI run、章节主张到工程文件/测试的映射、配图清单、验证结果和明确未实现项；每条主张还必须显式标为 `implemented`、`teaching_model` 或 `target_contract`，不能让读者从散落 prose 猜测 runtime 状态。以 Git 历史中包含该 manifest 的 commit 作为书稿 revision，避免在文件中自引用一个会因提交而变化的 SHA。

## 10. Definition of Done

一轮章节交付只有在以下条件同时成立时才完成：

- 正文已完整成章，而不是目录骨架；
- Go 示例与真实实现描述同一套状态和资金语义；
- 关键失败模式有测试或可运行验证；
- Mermaid、表格和正文配图有明确的信息作用；
- 支付语义、可靠性和发布审计问题已闭环；
- 单测、race、vet、契约、真实数据库 integration 与 Compose E2E 全绿；
- delivery manifest 已记录实现 SHA、测试证据和范围边界；
- 两个仓库已推送并核对远端 SHA；
- 清楚区分可运行支付核心、真实 PSP sandbox 接入和可处理真钱的 live 系统。
