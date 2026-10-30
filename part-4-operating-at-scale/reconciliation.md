# 第 17 章 对账

**中文** | [English](../en/part-4-operating-at-scale/reconciliation.md)

Aurora 的在线支付链路已经能够处理响应丢失：一笔 129.00 USD 的 SALE 超时后保持 UNKNOWN，恢复任务查询原操作，确认请款后再收敛。但月底核查仍发现三类异常：一笔平台显示失败的交易在服务商侧确实有请款记录；一笔已成功支付被运营人员在服务商后台退款；另一条服务商引用根本找不到对应的内部尝试。

在线恢复没有做错，它只是从来没有被赋予发现全部外部差异的能力。它知道哪些内部操作尚未完成，却不知道服务商侧有哪些平台未记录的操作；它恢复正在等待的工作，也不一定重新检查已经被内部判定为终态的记录。

本章的核心命题是：

> **对账是对多源事实进行持久化摄取、匹配、差异分类和受控修复的流程。运行批次、检查点、证据、差异、解决方案必须可追踪、可重放；对账发现矛盾，并不授予它绕过支付领域模型改写资金终态的权力。**

本章承接[第 15 章](observability.md)的观测与[第 16 章](experimentation.md)的结果口径。它不重写[第 10 章](../part-3-reliability/timeout-and-retry.md)的同操作恢复，也不负责账本调账；结算行（`SettlementLine`）、费用、FX 和实际成本留给[第 18 章](settlement.md)，跨区域对账所有权留给[第 19 章](../part-5-global-routing/multi-region.md)。

## 17.1 问题：一直等回调通知，为什么还是会有盲区

最初，Aurora 相信成功响应和回调通知足以更新状态。只要接口可用，回调迟一点也没关系。这个假设忽略了接收端停机、事件重复、投递耗尽、处理代码缺陷，以及服务商控制台中绕过平台发起的操作。

Stripe 官方明确说明回调通知不保证按生成顺序投递，接收方还必须处理重复事件；这些约束足以否定“每笔支付最终都会按顺序收到一次正确通知”的设计前提。[Stripe：回调通知](https://docs.stripe.com/webhooks) Adyen 的官方管理文档也展示了从客户专区（Customer Area）发起退款、取消或请款的能力；因此外部事实并不总由平台 API 请求产生。[Adyen：管理支付](https://docs.adyen.com/account/manage-payments)

发现这些差异后，最诱人的补丁是按年龄批量改状态：超过一天仍处理中，就认为失败；或者从运营提供的成功名单中直接 UPDATE 为成功。前一种做法把“没有答案”当成“没有副作用”，后一种做法绕过了账户、金额、引用与既有终态检查。两者都可能让一个排障动作制造新的资金错误。

这时系统需要的不是更大胆的修复脚本，而是能够回答：谁在什么范围内提供了什么证据，它与哪一个内部操作对应，哪里有差异，谁批准了什么动作，以及这个动作是否确实生效。

## 17.2 简单设计：先补查询，再看见查询的边界

第一步可以给超时支付增加后台查询，但不能把“超时太久”写成状态判定函数：

```go
package naive

import "time"

// WrongForFunds 有意展示不安全做法：经过多久并不是支付证据。
func WrongForFunds(updatedAt, now time.Time) string {
	if now.Sub(updatedAt) > 24*time.Hour {
		return "failed"
	}
	return "processing"
}
```

正确的在线恢复读取已持久化的操作，沿原连接、修订、服务商幂等键与语义载荷查询或恢复。明确证据到来前保持 UNKNOWN；它不因等待太久而切 PSP。这个能力必要，但仍是一个从内部待办出发的闭环。

| 维度 | 在线 UNKNOWN 恢复 | 周期对账 |
| --- | --- | --- |
| 出发点 | 已知、未完成的内部操作 | 固定范围内的内部记录与独立服务商清单 |
| 主要问题 | 这次操作现在能否安全收敛 | 两边有哪些缺失、重复、矛盾或人工变更 |
| 覆盖对象 | `created`、`sending`、处理中、未知等可恢复工作 | 按合同包含必要的终态与外部孤立记录 |
| 进度状态 | 每条操作的恢复时间、预算与租约 | 运行批次、分区检查点、来源覆盖水位 |
| 资金权限 | 已有恢复契约内的原操作动作 | 只提交证据与受控领域命令，不另造资金路径 |
| 完成标准 | 单条已知工作收敛或转复核 | 范围摄取、分类和处置进度分别可证明 |

只扫描内部 UNKNOWN，无论扫得多勤，都找不到“服务商有、内部没有”的交易；只导入服务商成功记录，也看不见“内部宣称成功、外部清单没有”的差异。因此周期对账需要双向视角：从内部已知对象检查外部证据，从服务商存量检查内部匹配。两边的范围、时间和完整性必须先对齐，不能把两个不同统计口径的集合直接相减。

## 17.3 规模化问题：比对之前，先让证据可靠

### 17.3.1 来源没有一条通用的优先级排行榜

“结算文件高于查询，查询高于回调通知”听起来明确，却混淆了不同事实。授权事件证明批准或额度保留，请款结果证明另一阶段，结算文件表达报告期间的资金或记账事项；它们并不是同一个字段的三个副本。

| 来源 | 可以提供的证据 | 不能直接推出 |
| --- | --- | --- |
| 同步响应 | 当次请求的特定服务商的结果 | 响应丢失等于未执行 |
| 已验签回调通知 | 指定事件或对象的归一化观察 | 到达更晚便有权覆盖所有旧事实 |
| 精确对象查询 | 在该 API 契约下读取的当前对象或操作状态 | 缓存、搜索索引与精确查询具有同等一致性 |
| 服务商交易提取 | 某个账户、时期、操作类别的记录集合 | 一个文件就是该账户全生命周期的完整清单 |
| 结算/费用文件 | 第 18 章定义的结算与经济事实 | 授权、请款、结算可以互换 |

每份证据保存来源类型、来源记录身份、解析器/API 版本、认证范围、有效时间、首次观察时间和必要的完整性信息。缺失来源时间时明确缺失，不用本地接收时间冒充。若两个可信来源矛盾，应保留双方，而不是给某种来源一个永久覆盖权。

“没有查到”尤其需要谨慎。Stripe PaymentIntent Search 官方说明它不适用于要求严格读后写一致性的流程，索引传播存在延迟。[Stripe：搜索 PaymentIntent](https://docs.stripe.com/api/payment_intents/search) 因而搜索为空、报表尚未到齐或查询超时，都不是原操作明确无副作用的证明。即使某个适配器的精确查询契约能给出确定未找到，是否允许同操作恢复仍由第 10 章的契约决定，对账工作进程不据此生成新尝试。

### 17.3.2 匹配首先是身份问题，不是金额相等问题

两笔 129.00 USD 的支付可以发生在同一分钟；商户也可能重用订单备注。同 PSP 的两个商户账户还可能有相同格式的引用。金额、币种和时间适合做一致性校验，却不应该独立构成自动认领交易的主键。

匹配先限定服务商账户与稳定命名空间，再使用被契约验证的操作关联或事务/操作引用，最后核对精确连接、修订、操作类型、金额和币种。只有一个候选且证明充分时才允许自动关联。多个候选、未知账户、引用冲突或身份漂移都产生差异，不采用“最接近时间的那条”。

| 匹配层 | 使用证据 | 结果规则 |
| --- | --- | --- |
| 强内部关联 | 经服务商契约验证的原操作 ID，并核对完整执行身份 | 唯一且属性一致才接受 |
| 强服务商关联 | 稳定命名空间内的支付/操作引用与操作类型 | 检查所有既有所有者，不能跨命名空间猜测 |
| 受控复合匹配 | 已审批的商户引用、账户、操作、金额币种与时间范围 | 通常生成调查候选；自动化需要单独证明唯一性 |
| 无匹配或多匹配 | 外部孤立记录、重复引用、资料不足 | 保存原证据，进入未匹配/含糊差异 |

Adyen 的报告文档分别定义 PSP 引用、商户引用与修改引用，并提醒合并报告时核对报告期间和时区。这说明父支付与后续请款/退款记录的引用不能机械混用，商户引用的唯一性也需要平台自己保证。[Adyen：准备报表](https://docs.adyen.com/platforms/prepare-reports/)

本轮代码中的 `provider_namespace_id` 是碰撞域，不是回调授权域。同命名空间的恢复别名不能因为名字不同绕过引用唯一性，也不能因为共享命名空间就认领另一条连接的操作。来源路径需要明确授权，证据持久化时还要回归原操作的规范身份。这个边界沿用第 3、11 章，本章只将它应用到比对流程。

### 17.3.3 扫过一个时间点，不等于完整覆盖了此前所有事实

第一版扫描器常把“最后一条记录时间”保存为检查点。若同一秒有多条记录，或者服务商后来补入一条更早发生的交易，下一轮就会漏掉数据。使用 OFFSET 也可能受集合插入、删除和排序变化影响。

一个运行批次应先冻结来源、账户范围、操作类别、时间区间、时区、过滤条件和适配器版本。上界固定，分页游标按该 API 的真实语义保存；不要跨来源复用一种假想的 `offset`。Stripe 的列表接口使用对象 ID 游标，而搜索接口使用另一种分页令牌，这种差异应该留在来源适配器中。[Stripe：分页](https://docs.stripe.com/api/pagination)

如果服务商提供快照令牌或完整文件清单，就记录它；若没有一致快照，则采用安全延迟、重叠回看和幂等摄取，并诚实标明最终一致的覆盖范围。仅按 `created_at` 扫描会漏掉旧对象的新修改，需要变更流、更新时间扫描、已知对象复查或周期宽窗口回看中的相应能力。重叠窗口只能覆盖其声明的迟到范围，不能假装解决任意久之后的更正。

扫描完成与结果全部一致是两个状态。一个运行批次可以完成摄取，却仍有未分类记录、待核实的外部缺失或开放复核。任何“缺失”判定都必须引用覆盖合同：账户、操作类型、时间区间和完整性水位；在数据尚未到齐时只能称为 `awaiting_evidence`。

### 17.3.4 重复记录和记录更正需要不同身份

回调通知事件、报告版本中的一行、一次查询观测，各有自己的来源身份。不能把可变服务商对象 ID 直接当作不可变事件 ID：同一对象今天处理中、明天已请款，是两次合法观察，不是载荷篡改。查询每次生成新的观测身份；同一观察的重复投递则复用原 ID。

对于真正不可变的来源记录，以“来源范围 + 记录 ID + 记录修订”建立唯一约束；同键同摘要是重复，同键不同摘要应隔离并保留两份内容。服务商发布了正式更正版本，就追加新证据并关联被更正记录，不能覆写旧文件或旧规范化结果。解析器修复也生成新的解释版本，原始最小化证据保持可追溯。

这些记录不能成为新的敏感数据仓库。源文件或 API 返回可能包含顾客资料、支付凭证或秘密字段；摄取边界先做允许列表，凭证不进入证据、日志、追踪或事件。必要的来源原件按受控数据类别保存与授权访问，而不是把整个响应体随意塞进 JSONB。

## 17.4 生产级设计：把对账变成持久工作流

### 17.4.1 五类对象，各自回答一个问题

```mermaid
flowchart LR
    S["服务商查询 / 数据提取<br/>已验证来源范围"] --> I["摄取并归一化"]
    I --> E["证据<br/>只追加事实"]
    R["运行批次 + 检查点<br/>有界来源覆盖"] --> I
    E --> M["与支付服务持有的快照匹配"]
    M --> D["差异<br/>已分类问题与证据链接"]
    D --> P["解决建议<br/>稳定命令与前置条件"]
    P --> A{"允许的单调修复?"}
    A -->|"是"| C["支付领域命令"]
    A -->|"否 / 存在冲突"| Q["人工复核或隔离区"]
    C --> O["回执 + 解决历史"]
    Q --> O
```

| 对象 | 核心字段 | 持久化不变量 |
| --- | --- | --- |
| 运行批次 | 运行批次 ID、来源/作用域、固定窗口、模式/适配器版本、快照引用、租约代数、状态 | 重启不改变原范围；新范围产生新运行批次 |
| 检查点 | 运行批次/分区、不透明游标、检查点版本、摄取水位 | 只推进到已持久接管的页，旧所有者不能覆盖新进度 |
| 证据 | 证据 ID、来源身份/修订、摘要、规范连接、操作/引用、金额币种、时间 | 原件追加；重复幂等；冲突隔离；解释版本独立 |
| 差异 | 差异 ID、类型、主体、证据集合、分类器版本、当前处置状态 | 不因重跑同一问题产生大量重复工单；历史分类不抹除 |
| 解决方案 | 解决方案/命令 ID、差异 ID、动作、证据引用、预期领域版本、批准者、执行回执 | 提议不是执行；回执不是猜测；每次处置留审计记录 |

差异的身份通常围绕一个稳定问题，例如某操作的引用冲突；多轮运行发现相同问题时追加观测链接，而不是按 `run_id` 每次新建一张工单。问题解决后出现新矛盾，可以显式重新打开或产生关联的新案例，不能覆盖旧解决方案。记录这些身份是为了恢复工作，不是把它们放进高基数指标标签。

首个部署可以是 `services/payment/cmd/reconciler`：独立进程和资源预算，但仍在支付数据所有权内。它通过网关的只读契约取得外部观察，不读取网关模式；若以后成为独立服务，则用支付 API/事件快照匹配并提交命令，不持有跨库任意 UPDATE 权限。

### 17.4.2 检查点必须与“已接管的工作”原子提交

最关键的崩溃窗口位于取回一页与保存进度之间。先更新游标、再写证据，崩溃后会永久跳过这一页；先逐条修复支付、最后才记录工作，崩溃后又无法证明哪些动作已经生效。

正确边界是一个短本地事务：保存该页的证据或隔离项、为每条可处理记录创建幂等匹配工作、推进检查点，并记录页完成信息。网络抓取发生在事务之外，分类与资金状态收敛通过后续持久任务完成。检查点代表“工作已经可靠接管”，不代表“所有差异已经修复”。

```mermaid
sequenceDiagram
    participant W as 对账工作进程
    participant S as 来源适配器
    participant DB as 由支付服务持有的对账数据库
    participant M as 匹配器 / 解决处理器
    W->>DB: 读取运行批次、租约代数和检查点 v7
    W->>S: 从游标 C7 获取固定作用域
    S-->>W: 返回当前页和下一游标 C8
    W->>DB: 用当前租约和检查点护栏开始事务
    W->>DB: 追加证据 / 隔离记录和持久工作项
    W->>DB: 将检查点从 v7 推进到 v8
    DB-->>W: 提交
    Note over W,DB: 提交前崩溃可安全地重新获取 C7
    M->>DB: 领取持久化匹配工作
    M->>M: 匹配、分类并提出受控解决建议
    Note over M,DB: 游标进度不依赖所有案例都已解决
```

下面的标准库 Go 示例表达页级交接，不提供内存数据库冒充持久化。`CommitPage` 必须由真实 PostgreSQL 事务实现，注释中的原子性要由集成测试兑现。

```go
package reconciliation

import (
	"context"
	"errors"
)

type Run struct {
	ID             string
	ScopeDigest    string
	AdapterVersion string
	LeaseEpoch     uint64
}

type Checkpoint struct {
	RunID   string
	Version uint64
	Cursor  string
}

type SourceRecord struct {
	RecordID       string
	RecordRevision string
	PayloadDigest  [32]byte
	SanitizedBody  []byte
}

type Page struct {
	ScopeDigest    string
	AdapterVersion string
	FromCursor     string
	NextCursor     string
	Last           bool
	Records        []SourceRecord
}

type Source interface {
	Fetch(context.Context, Run, string) (Page, error)
}

type PageStore interface {
		// 单个事务验证当前租约代数和检查点版本，
		// 持久保存每条记录或隔离项及幂等匹配工作，
		// 再推进检查点；它绝不调用服务商资金接口。
	CommitPage(context.Context, Run, Checkpoint, Page) (Checkpoint, error)
}

func ScanPage(ctx context.Context, source Source, store PageStore, run Run, current Checkpoint) (Checkpoint, error) {
	if run.ID == "" || run.ScopeDigest == "" || run.AdapterVersion == "" ||
		run.LeaseEpoch == 0 || current.RunID != run.ID {
		return current, errors.New("invalid run or checkpoint")
	}
	page, err := source.Fetch(ctx, run, current.Cursor)
	if err != nil {
		return current, err
	}
	if page.ScopeDigest != run.ScopeDigest || page.AdapterVersion != run.AdapterVersion ||
		page.FromCursor != current.Cursor ||
		(!page.Last && (page.NextCursor == "" || page.NextCursor == current.Cursor)) {
		return current, errors.New("page does not match fixed scan contract")
	}
	return store.CommitPage(ctx, run, current, page)
}
```

`ScopeDigest` 引用冻结的账户、操作、窗口和过滤条件，不由临时请求拼接。存储还需防止游标循环、重复页产生重复工作，以及已失去租约的旧工作进程提交。单纯“领取后等一分钟”不提供栅栏保护；每次提交都要验证当前代数与检查点版本。

`FOR UPDATE SKIP LOCKED` 可用于分配待处理任务，但不能用“本次没取到行”证明整个范围为空。PostgreSQL 官方明确指出跳过锁定行会产生不一致视图，因此它适合队列式消费，不适合当作一般完整性扫描的证明。[PostgreSQL：SELECT](https://www.postgresql.org/docs/current/sql-select.html) 范围完成应由全部分区检查点和明确的覆盖元数据判断。

### 17.4.3 分类器只提出动作，不执行动作

差异类型与修复动作应当分开保存。同一种 `status_mismatch`，可能只是未知状态被充分证据补齐，也可能是两个终态互相矛盾；名字相同不代表有相同的自动化权限。

| 差异类型 | 产生条件 | 默认处置边界 |
| --- | --- | --- |
| 内部缺失（`missing_internal`） | 可信外部记录没有唯一内部对象 | 调查来源和关联，不自动创建支付 |
| 外部缺失（`missing_external`） | 声明的完整范围中缺少预期外部记录 | 复查范围与精确对象；缺失本身不证明无副作用 |
| 状态不匹配（`status_mismatch`） | 两侧同一操作的阶段或状态不一致 | 仅充分且无冲突的新事实可提出单调收敛 |
| 金额或币种不匹配（`amount_currency_mismatch`） | 强身份关联后金额或币种矛盾 | 复核，不自动补差额或创建退款 |
| 引用冲突（`reference_collision`） | 同一碰撞域内出现竞争所有者 | 保留全部关联并复核，不抢占引用 |
| 等待证据（`awaiting_evidence`） | 覆盖未成熟或关键属性缺失 | 保留待办和年龄，不伪装成已一致 |

不可信或无法验证来源身份的输入先进入隔离区，不能仅凭外部声称的操作 ID 给某笔支付贴上冲突标签。通过信任边界但尚未匹配的记录则可形成独立差异，等调查取得强关联后再涉及支付。

下面的纯函数只覆盖“已有内部操作 ID 的自动请款 SALE”这条窄路径。基于引用的候选发现与唯一性检查在它之前完成；无法获得强关联的证据保留在未匹配队列。`Verified` 由受信任的网关/来源适配器设置，不是外部调用者可自行声明的权限字段。

```go
package reconciliation

type Connection struct {
	ID        string
	Revision  string
	Namespace string
	Provider  string
}

type Snapshot struct {
	OperationID        string
	Connection         Connection
	Version            int64
	Status             string
	Kind               string
	AmountMinor        int64
	Currency           string
	PaymentReference   string
	OperationReference string
	ReviewRequired     bool
	AuthorizationSeen  bool
}

type Evidence struct {
	ID                 string
	OperationID        string
	Connection         Connection
	Kind               string
	State              string
	Verified           bool
	AmountKnown        bool
	AmountMinor        int64
	Currency           string
	PaymentReference   string
	OperationReference string
	NoSideEffect       bool
	ReferenceCollision bool
}

type Plan struct {
	Action          string
	Reason          string
	EvidenceID      string
	ExpectedVersion int64
}

func Classify(local Snapshot, evidence Evidence) Plan {
	plan := func(action, reason string) Plan {
		return Plan{action, reason, evidence.ID, local.Version}
	}
	if !evidence.Verified || evidence.ID == "" || local.Version <= 0 ||
		local.OperationID == "" || evidence.OperationID != local.OperationID ||
		local.Connection.ID == "" || local.Connection.Revision == "" ||
		local.Connection.Namespace == "" || local.Connection.Provider == "" ||
		evidence.Connection != local.Connection {
		return plan("quarantine", "untrusted_or_unmatched_identity")
	}
	if local.Kind != "sale" || evidence.Kind != local.Kind {
		return plan("review", "unsupported_operation_mapping")
	}
	if local.ReviewRequired {
		return plan("review", "existing_review_is_sticky")
	}
	if evidence.ReferenceCollision ||
		(local.PaymentReference != "" && evidence.PaymentReference != "" &&
			local.PaymentReference != evidence.PaymentReference) ||
		(local.OperationReference != "" && evidence.OperationReference != "" &&
			local.OperationReference != evidence.OperationReference) {
		return plan("review", "reference_conflict")
	}
	if local.AmountMinor <= 0 || local.Currency == "" ||
		(evidence.AmountKnown && (evidence.AmountMinor != local.AmountMinor ||
			evidence.Currency != local.Currency)) {
		return plan("review", "amount_or_currency_mismatch")
	}
	terminal := local.Status == "succeeded" || local.Status == "failed"
	switch evidence.State {
	case "unknown", "pending", "not_found":
		if terminal {
			return plan("no_op", "retain_terminal_against_weaker_evidence")
		}
		return plan("wait", "more_evidence_required")
	case "authorized":
		if local.Status == "failed" {
			return plan("review", "authorization_conflicts_with_failure")
		}
		if local.Status == "succeeded" {
			return plan("no_op", "authorization_is_not_new_capture_evidence")
		}
		return plan("wait", "authorization_is_not_capture")
	case "captured", "declined", "failed":
	default:
		return plan("review", "unsupported_source_state")
	}
	if evidence.State == "captured" && local.Status == "failed" {
		return plan("review", "terminal_status_conflict")
	}
	if evidence.State != "captured" && evidence.NoSideEffect &&
		(local.Status == "succeeded" || local.AuthorizationSeen) {
		return plan("review", "funds_evidence_conflicts_with_failure")
	}
	if !evidence.AmountKnown {
		return plan("wait", "amount_evidence_missing")
	}
	if evidence.State == "captured" {
		if evidence.PaymentReference == "" || evidence.OperationReference == "" {
			return plan("review", "capture_reference_missing")
		}
	} else {
		if !evidence.NoSideEffect {
			return plan("wait", "failure_does_not_prove_no_side_effect")
		}
	}
	if terminal {
		return plan("no_op", "consistent_terminal_evidence")
	}
	switch local.Status {
	case "created", "sending", "pending", "unknown", "retry_scheduled":
		return plan("apply_outcome", "monotonic_evidence_candidate")
	default:
		return plan("review", "unsupported_local_state")
	}
}
```

这段函数不会把授权当请款，也不会把任意失败字符串当作 `side_effect=none`。金额缺失与金额为零不同；已经显露的终态矛盾也不能因为金额缺失而降级为普通等待。部分请款、退款或不支持的操作必须由对应模型解释，不能硬塞进全额 SALE 的成功分支。它返回 `apply_outcome` 也只是候选动作，最终仍由支付在最新事务状态下重新校验。

这里的 `no_op` 只表示不改变资金状态，并不删除证据或自动关闭全部差异。例如内部已成功而搜索暂未查到，原终态必须保留，但来源覆盖问题仍然可以继续调查。

### 17.4.4 自动修复必须经过支付的原子边界

匹配器读取快照 v8 后，回调通知可能已经把操作更新为 v9。若它拿旧比较结果直接 UPDATE，就会覆盖刚到达的事实。解决方案命令因此必须携带稳定命令 ID、差异/证据引用、预期领域版本和允许的动作类别。

支付在本地事务内验证调用者权限、幂等命令身份、证据绑定与当前聚合，调用同一套 `ApplyProviderOutcome` 或 `RequireManualReview`，持久化回执并写状态发件箱。版本改变时重新读取与分类；不能简单取消版本条件。PostgreSQL 的事务隔离并不会自动替应用维护所有业务约束，需要正确的锁、版本检查与必要的整事务重试。[PostgreSQL：事务隔离](https://www.postgresql.org/docs/current/transaction-iso.html)

```mermaid
sequenceDiagram
    participant R as 解决处理器
    participant D as 差异 / 解决记录存储
    participant P as 支付命令处理器
    participant DB as 支付数据库
    R->>D: 持久化建议和稳定命令 ID
    R->>P: 证据引用 + 预期版本 + 允许动作
    P->>DB: 锁定聚合并检查命令收件箱
    alt 新命令，前置条件仍有效
        P->>DB: 执行领域迁移或要求复核
        P->>DB: 提交命令回执和状态发件箱事件
    else 重复命令
        DB-->>P: 返回已有持久化回执
    else 状态已变化
        DB-->>P: 需要重新分类
    end
    P-->>R: 持久化回执或前置条件过期结果
    R->>D: 追加执行结果，仅在证据充分时关闭
    Note over R,DB: 确认丢失时重试同一命令，绝不创建新资金操作
```

若命令已生效而回执响应丢失，解决处理器继续发送同一个命令 ID，支付返回已保存回执。若保存解决方案结果时崩溃，也通过这份回执补齐，而不是再次“修复资金”。唯一约束应覆盖命令身份与来源事实身份；数据库唯一索引是并发去重的保护，不是先查询再插入的乐观约定。[PostgreSQL：唯一索引](https://www.postgresql.org/docs/current/indexes-unique.html)

这里的修复命令 ID 不是商户键、服务商幂等键，也不应伪装成新的服务商事件 ID。它属于支付受控修复接口的幂等命名空间；真实来源身份继续保留在证据中。自动修复最多补入已经发生且被验证的事实，不发起新的 SALE、请款、撤销或退款。

### 17.4.5 冲突复核是正式出口，不是失败消息的垃圾桶

内部失败与外部已请款矛盾时，系统保存双方证据并标记完整性复核；不能为了让日报一致就覆盖成成功。内部已确认成功也不能被迟到的失败覆盖。对未终结且无法安全归属的操作，可以停止自动恢复并进入 `review_required`；对已确认终态，则保持原资金事实并叠加完整性问题。

人工队列需要所有者、优先级、年龄、调查步骤和可审计处置。操作者可以补充证据、确认来源映射错误、将问题转给服务商、提出受控领域修复，或交由外部账本流程处理会计调整；不能直接在对账表写一个“余额已修正”。需要退款等新的资金操作时，它是另一个明确授权、幂等且绑定原连接/引用的工作流，不是关闭差异的副作用。

未知内部对象的服务商记录先成为 `missing_internal`，不能自动创建假的支付意图来完成关联。引用被两个尝试声称时，应保留两个所有者的关系并让相关资金对象进入复核，而不是把引用抢给最后到达的一方。

## 17.5 故障模式：用崩溃与矛盾证明设计

| 失败场景 | 必须保留的事实 | 正确结果 |
| --- | --- | --- |
| 回调通知丢失，查询确认原 SALE 已请款 | 原操作、精确身份、金额币种和必要引用 | 通过支付领域命令收敛，只有原资金路径 |
| 内部失败，外部已请款 | 两个互相矛盾的观察及来源时间 | 复核，不覆盖终态、不发起另一笔支付 |
| 来源搜索为空或报告未到齐 | 查询范围、覆盖水位与来源质量 | 等待证据（`awaiting_evidence`），不推出无副作用 |
| 同命名空间的引用被两条连接声称 | 全部所有权主张、规范身份与传入证据 | 引用冲突，不能借别名绕过检测 |
| 页提交前崩溃 | 旧检查点 | 重取同页，记录幂等，不漏工作 |
| 页提交后、分类前崩溃 | 新检查点、证据和持久化工作 | 从任务表继续，不依赖进程内列表 |
| 修复提交后 ACK 丢失 | 同一个命令 ID 与支付持久化回执 | 幂等返回，不产生第二次领域动作 |
| 老工作进程在租约失效后提交 | 新代数、检查点版本 | 栅栏机制拒绝旧所有者，不回退游标 |
| 更正文件或解析器升级 | 原版本、新版本与替代/解释关系 | 追加解释，不删除原证据 |
| 终态复查发现外部退款 | 原请款与独立修改证据 | 按支持的操作模型或复核处理，不把原请款改成从未成功 |

纯函数测试可以先证明“分类不等于资金变更”。以下用例覆盖 UNKNOWN 收敛候选、矛盾终态、授权未请款、含糊失败以及命名空间漂移。它们不代替真实数据库的事务和崩溃窗口测试。

```go
package reconciliation

import "testing"

func TestClassifyNeverOverwritesFunds(t *testing.T) {
	connection := Connection{"cand_us_a", "rev7", "account_us_a", "psp_a"}
	base := Snapshot{
		OperationID: "op_1", Connection: connection, Version: 8,
		Status: "unknown", Kind: "sale", AmountMinor: 12900, Currency: "USD",
	}
	proof := Evidence{
		ID: "ev_1", OperationID: "op_1", Connection: connection,
		Kind: "sale", State: "captured", Verified: true, AmountKnown: true,
		AmountMinor: 12900, Currency: "USD",
		PaymentReference: "payment_ref_1", OperationReference: "sale_ref_1",
	}
	tests := []struct {
		name   string
		change func(*Snapshot, *Evidence)
		action string
	}{
		{"unknown to captured candidate", func(*Snapshot, *Evidence) {}, "apply_outcome"},
		{"failed conflicts with captured", func(s *Snapshot, _ *Evidence) { s.Status = "failed" }, "review"},
		{"missing amount cannot hide terminal conflict", func(s *Snapshot, e *Evidence) {
			s.Status, e.AmountKnown = "failed", false
		}, "review"},
		{"success conflicts with conclusive failure even without amount", func(s *Snapshot, e *Evidence) {
			s.Status, e.State, e.NoSideEffect, e.AmountKnown = "succeeded", "failed", true, false
		}, "review"},
		{"authorization is not capture", func(_ *Snapshot, e *Evidence) { e.State = "authorized" }, "wait"},
		{"ambiguous failure stays unresolved", func(_ *Snapshot, e *Evidence) { e.State = "failed" }, "wait"},
		{"foreign namespace cannot attach", func(_ *Snapshot, e *Evidence) { e.Connection.Namespace = "other_account" }, "quarantine"},
		{"reference collision", func(_ *Snapshot, e *Evidence) { e.ReferenceCollision = true }, "review"},
		{"amount mismatch", func(_ *Snapshot, e *Evidence) { e.AmountMinor++ }, "review"},
		{"search miss is not failure", func(_ *Snapshot, e *Evidence) { e.State = "not_found" }, "wait"},
		{"existing review stays sticky", func(s *Snapshot, _ *Evidence) { s.ReviewRequired = true }, "review"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			local, evidence := base, proof
			test.change(&local, &evidence)
			before := local
			plan := Classify(local, evidence)
			if plan.Action != test.action || plan.ExpectedVersion != local.Version {
				t.Fatalf("plan=%+v, want action %s", plan, test.action)
			}
			if local != before {
				t.Fatal("classifier mutated payment snapshot")
			}
		})
	}
}
```

工程验收还必须在真实 PostgreSQL 中测试同页并发摄取、同键不同摘要、检查点 CAS、旧租约代数和命令回执丢失。测试应同时检查证据数量、开放差异、解决方案回执、领域状态和资金调用次数；只断言一个函数返回 `nil`，无法证明没有丢记录或重复动作。

运营上分别监控来源覆盖落后、最老未分类任务、开放差异年龄、复核积压和自动修复拒绝原因。运行批次的“绿色”必须说明是摄取完成、分类完成还是差异已处置；有一千条人工问题未关闭的任务不能只报一句“对账成功”。

## 17.6 权衡与配套实现边界

频繁精确查询能更快发现变化，却消耗服务商限额并可能挤占在线恢复；批量提取适合广覆盖，但具有来源延迟和不同报告语义。合理设计保留独立并发预算与优先级，让周期扫描可退让，不能在大规模回看时饿死正在等待答案的支付。

宽窗口回看提高迟到覆盖，却增加重复摄取；幂等证据与稳定差异能降低重复工作的代价，但不能代替明确的迟到上限。完整保留来源有助于审计，数据最小化与保留要求又限制可存内容；因此保留版本、摘要和受控原件引用，不把全部服务商数据复制到路由服务或日志。

自动修复能降低人工量，却必须保持狭窄条件。若为了提高“自动解决率”放宽引用匹配或忽略资金冲突，系统只是把可见差异变成不可见错误。宁可留下可解释的复核，也不要获得一张无法说明来历的平衡报表。

当前 `YouEclipse/payments` 的实现与本章目标需要明确分开：

| 工程位置 | 当前已经具备 | 仍是待配套增量 |
| --- | --- | --- |
| 支付 `app.RunReconciler` | 轮询已知可恢复操作，领取后查询或按现有契约恢复原操作 | 独立周期运行批次、双向存量、终态复查 |
| `ClaimRecoverableOperations` | PostgreSQL 中按恢复时间领取 `created`、`sending`、处理中、未知等工作 | 来源分区检查点、固定扫描范围、租约代数栅栏 |
| 网关 `Get(..., reconcile=true)` | 对适用的未终态查询原连接；终态直接返回本地记录 | 对终态仍可取得新外部事实的显式只读审计查询/提取 API |
| 网关 `provider_operation_observations` | 销售操作/查询/回调通知与旧版种子观察以来源、内容摘要、时间和精确操作身份追加保存；数据库阻止 UPDATE/DELETE/TRUNCATE | 跨来源存量、范围水位、修订链和面向对账的读取 API |
| 支付 `ApplyOutcome` | 收件箱、身份校验、状态转换与发件箱的持久事务；拒绝数据进入隔离区 | 面向对账的证据/差异/解决方案模型与修复命令回执 |
| 领域 `ApplyProviderOutcome` / `RequireManualReview` | 单调收敛、保留确认终态、冲突完整性标记；停止不安全自动恢复 | 完整人工调查与已审批修复流程 |
| 周期对账服务 | 尚不存在独立实现 | `services/payment/cmd/reconciler`、`internal/reconciliation` 及迁移、失败测试 |

特别是现有网关 GET 会对成功、失败和 `review_required` 短路返回；循环调用它并不能独立验证这些终态。新审计接口应明确只读取服务商，不执行重送，不把 `operation_resubmit_required` 当作对账工作进程的执行授权。网关继续拥有凭证、特定服务商的查询与归一化，支付继续拥有资金状态。

当前基线已有精确连接/修订/命名空间、UNKNOWN 查询收敛、终态冲突与引用冲突测试，也新增了网关侧的追加式原始观察账本。该账本解决“协议观察不能被后来的状态覆盖”，却不等于本章独立对账服务已经交付：它没有冻结外部存量范围，没有运行批次/检查点/差异/解决方案，也没有修复命令的崩溃窗口。不能把观测行或教学分类器冒充完整平台。

建议的首个纵向切片只覆盖一个服务商模拟器的全额自动请款 SALE：通过网关只读观察导入固定窗口，匹配原操作，持久化缺失/状态/引用差异，再由支付受控命令收敛或复核。能识别外部退款记录，不等于当前系统已经提供人工请款、撤销、退款或账本调账 API。

## 17.7 小结

本章从“等回调、补查询”走到一条持久的事实比对链：冻结运行范围，原子接管分页工作，追加证据，先做强身份匹配，再分类差异，最后通过有前置条件和幂等回执的解决方案进入支付领域边界。

时间过去、搜索为空、记录晚到或报表出现一行，都不足以自动改写资金终态。UNKNOWN 可以在充分证据下单调收敛；矛盾必须保留并进入复核；任何新资金操作需要自己的授权与幂等契约。周期对账扩大的是发现事实的范围，不是系统任意修改事实的权力。下一章将在这条证据链上归因结算与实际费用，而不是再创造一套“看起来对得上”的支付真相。
