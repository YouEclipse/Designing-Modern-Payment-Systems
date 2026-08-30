# 第 6 章《服务商准入筛选》交付记录

## 章节契约

- 核心命题：进入排序集合必须由连接级、带来源与时效的证据证明，是一项硬性准入条件；任何偏好、成本或实验都不能越过它。
- 演进线：支付服务商品牌布尔值 → 账户、区域与令牌差异 → 多维三值谓词 → 不可变能力目录与“证据不足即拒绝”的准入筛选。
- 本章增量：`capability_catalog.v1` 编译器与准入筛选器、完整候选通道清单、描述符来源与新鲜度，以及逐候选通道的准入证据。
- 后续范围：金额区间、商户合同、合规、健康度、费用与成功率分别留给后续能力目录版本及第 7、8、13 章。

## 能力状态

- `implemented`：固定实现中的 `capability_catalog.v1` 编译与准入筛选、精确连接修订、三值事实、新鲜度与逐候选证据。
- `teaching_model`：正文 5 个 Go 代码块中的旧版筛选器与能力目录最小模型。
- `target_contract`：金额、商户与合规约束、在线服务商发现、动态健康度、真实账户合同与令牌可移植性。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 纵向切片：路由服务 v2 按精确连接修订，对币种、国家、支付方式、资金操作、请款模式、凭证域与时间做准入筛选；规则只接收已准入的候选通道编号。

## 主张—实现—测试

工程文件相对于固定实现。

| 状态 | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | 能力属于精确连接修订，不属于 PSP 品牌 | 能力制品与目录 | `TestCatalogEligibilityUsesExactConnectionAndThreeValues` |
| `implemented` | 半快照、别名与含糊制品不能启用 | 能力编译器 | `TestCatalogRejectsPartialAndAmbiguousArtifacts` |
| `implemented` | 未知或过期时一律拒绝，已知不支持时保留解释 | 能力准入筛选器与决策证据 | 能力目录单元测试；`TestV2UnknownCredentialIsExplicitNoCandidate` |
| `implemented` | 币种与凭证的硬约束先于规则 | 决策流水线 | `TestPipelineForceAndSplitNeverOverrideCredentialOrCurrencyEligibility` |
| `implemented` | 快照可重放，但过期后不可用于新执行 | 目录与规则有效期、HTTP 准入 | 重放与在线执行对照测试 |
| `implemented` | 无候选通道时返回证据且不触发网关 | 路由服务 v2 HTTP；支付客户端 | `422` 证据测试；Compose 无合格候选路径 |
| `implemented` | 选中连接与执行连接保持同一身份 | 路由决策、支付尝试、资金操作、服务商网关 | PostgreSQL 集成测试与 Compose 元组断言 |

## 图表与配图

- 4 个 Mermaid 已实际渲染，覆盖硬约束顺序、多维谓词、快照刷新与适配器、网关边界。
- Ian 正文配图 `assets/provider-eligibility-illustrations/01-same-brand-different-doors.png` 为 1672×941，表达“同一 PSP 品牌不等于同一扇门”，尺寸、路径和正文嵌入已检查。
- 本章 5 个 Go 代码块由正文提取为 2 个独立包，编译、竞态检测和静态检查通过；附加测试覆盖新鲜度、边界和不可复活约束。

## 发布验证

固定实现的单元测试、竞态检测、静态检查、OpenAPI、真实 PostgreSQL 集成测试、服务镜像构建和 Compose 端到端测试均在持续集成任务 `33041147119` 中成功；书稿本地链接、Mermaid、图片和固定实现链接通过检查。代码远端提交值已核对；书稿版本由包含本记录的 Git 历史确定。

## 能力边界

当前六组线格式能力不包括金额范围、商户允许清单、合规标签、健康度或在线服务商发现。本地能力目录是两个模拟器连接的演示事实，不证明真实 PSP 账户、合同、令牌可移植性或监管资格；配图中的多个“门”不是多个真实生产 PSP。系统没有真实 PSP 沙箱或生产支付能力。
