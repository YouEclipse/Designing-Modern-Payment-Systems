# 第 6 章《Provider Eligibility》交付记录

## 章节契约

- 核心命题：进入排序集合是一项必须由 connection-level、带来源与时效的证据证明的硬资格；任何偏好、成本或实验都不能越过它。
- 演进线：Provider 品牌布尔值 → 账户/Region/Token 差异 → 多维三值谓词 → 不可变 Capability Catalog 与 fail-closed 过滤。
- 本章增量：`capability_catalog.v1` compiler/filter、完整 Candidate manifest、Descriptor provenance/freshness、逐候选 Eligibility evidence。
- 后续范围：金额区间、Merchant 合同、Compliance、Health、费用与成功率分别留给后续 Catalog 版本及第 7、8、13 章。

## 配套实现

- 固定实现：[6583e4dc37b391a8ecbe51830780717f1c0f8843](https://github.com/YouEclipse/payments/commit/6583e4dc37b391a8ecbe51830780717f1c0f8843)。
- CI：[32987908560 — success](https://github.com/YouEclipse/payments/actions/runs/32987908560)。
- 纵向切片：Router v2 按精确 connection revision 过滤 Currency、country、Payment Method、Operation/Capture、credential domain 与时间；Rule 只接收 eligible IDs。

## 主张—实现—测试

工程文件相对于固定实现。

| 主张 | 工程文件 | 验证 |
| --- | --- | --- |
| 能力属于精确 connection revision，不属于 PSP 品牌 | capability Artifact/Catalog | `TestCatalogEligibilityUsesExactConnectionAndThreeValues` |
| 半快照、别名与含糊 Artifact 不能启用 | capability compiler | `TestCatalogRejectsPartialAndAmbiguousArtifacts` |
| unknown/stale fail closed，已知 no 保留解释 | capability filter + decision evidence | Catalog unit tests；`TestV2UnknownCredentialIsExplicitNoCandidate` |
| Currency/credential 硬约束先于 Rule | decision pipeline | `TestPipelineForceAndSplitNeverOverrideCredentialOrCurrencyEligibility` |
| Snapshot 可重放但过期后不可用于新执行 | catalog/rules validity + HTTP admission | replay/online execution contrast tests |
| 无候选返回证据且不触发 Gateway | Router v2 HTTP；Payment client | `422` evidence tests；Compose no-eligible path |
| 选中连接与执行连接保持同一身份 | Decision、Payment Attempt/Operation、Gateway | PostgreSQL integration 与 Compose tuple assertions |

## 图表与配图

- 4 个 Mermaid 已实际渲染，覆盖硬约束顺序、多维谓词、snapshot refresh 与 Adapter/Gateway 边界。
- Ian 正文配图 `assets/provider-eligibility-illustrations/01-same-brand-different-doors.png` 为 1672×941，表达“同一 PSP 品牌不等于同一扇门”，尺寸、路径和正文嵌入已检查。
- 本章 5 个 Go 块由正文提取为 2 个独立 package，compile/race/vet 通过；附加测试覆盖 freshness、边界和不可复活约束。

## 发布验证

固定实现的 unit、race、vet、OpenAPI、真实 PostgreSQL integration、服务镜像构建和 Compose E2E 均在 CI run `32987908560` 成功；书稿本地链接、Mermaid、图片和固定实现链接通过检查。代码远端 SHA 已核对；书稿 revision 由包含本记录的 Git 历史确定。

## 能力边界

当前六组 wire 能力不包括金额范围、Merchant allowlist、Compliance Tags、Health 或在线 Provider discovery。Local Catalog 是两个 Simulator connection 的演示事实，不证明真实 PSP 账户、合同、Token portability 或监管资格；配图中的多个“门”不是多个 live PSP。没有真实 PSP sandbox 或 live payment 能力。
