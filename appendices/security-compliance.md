# 附录 B 安全、合规与数据边界

一次线上排查很容易从“为什么这笔支付没走通”变成“把整个请求打出来看看”。请求被复制进日志，日志进入检索平台，错误工单又把它贴进聊天群。原本只应该在凭证边界短暂出现的数据，就这样获得了比支付业务更广的访问者和更长的寿命。增加 TLS 或把服务拆成微服务，并不会撤销这些副本。

本附录的核心命题是：安全首先是限制数据和权限的传播，其次才是保护已经允许存在的数据。它把[第 3 章的服务边界](../part-1-foundations/routing-system-boundaries.md)、[第 4 章的上下文](../part-2-routing-engine/routing-context.md)、[第 14 章的决策记录](../part-4-operating-at-scale/routing-decision-store.md)和[第 20 章的配置治理](../part-5-global-routing/configuration-platform.md)串成横向检查方法，不另建一套反欺诈引擎，也不代替目标市场的安全评估、合同审核和法律意见。

## B.1 先区分“本系统不保存”与“标准一律禁止保存”

PAN 属于持卡人数据 CHD；CVV/CVC/CID、完整磁道和 PIN/PIN Block 属于敏感认证数据 SAD。两者不能混称。对本书讨论的商户收单路由场景，SAD 在授权完成后不得保存，即使加密也不例外；PAN 则有不同的存储保护与业务必要性要求。本项目采取更窄的架构政策：原始 PAN 和 SAD 均不进入平台数据库、路由服务、决策、日志或消息；凭证采集交给经过评估的托管组件或密钥库。[PCI SSC：CVV 与授权后存储要求](https://www.pcisecuritystandards.org/faqs/are-merchants-allowed-to-request-card-verification-codes-values-from-cardholders/)

令牌不是一个统一安全等级。服务商密钥库令牌、平台受控引用、网络支付令牌可能具有不同的可逆性、用途和域限制。一个服务拿不到 PAN，却能任意调用去令牌化 API，不能因此自称已脱离持卡人数据环境。EMVCo 支付令牌也有特定定义和域控制；PCI SSC 明确区分令牌本身与同时处理账户数据、或连接相关系统的环境。不能把这一结论推广为“所有叫令牌的字段都免于 PCI”。[PCI SSC：EMVCo 支付令牌的作用域](https://www.pcisecuritystandards.org/faqs/1326/)

PCI DSS 是技术与运行控制基线，评估对象还包括能够影响持卡人数据环境（CDE）安全的组件。实际责任要结合接入方式、网络隔离、服务商责任矩阵和评估范围确认；本书不宣布已经通过 PCI 认证，也不提供一个可通用于所有商户的 SAQ 结论。[PCI SSC：PCI DSS 适用对象与评估](https://www.pcisecuritystandards.org/standards/pci-dss/)

| 数据类别 | 合理持有者 | 本项目的持久化边界 | 路由服务/通用观测平台 |
| --- | --- | --- | --- |
| 原始 PAN、SAD | 经评估的支付凭证采集边界 | 平台不接收和落盘；SAD 不因调试例外保存 | 禁止 |
| 受控支付方式引用 | 支付服务，执行时交给获授权支付网关 | 按用途、访问控制和保留策略保护 | 不发送令牌，只传必要派生属性 |
| 服务商 API 凭证/回调秘密 | 指定连接的支付网关或密钥服务 | 秘密存储；不得写入业务事件和配置摘要 | 禁止 |
| 金额、币种、必要国家/卡网络属性 | 支付服务与路由服务 | 使用版本化、最小化字段 | 允许必要字段，不附带原请求 |
| 服务商引用、内部支付标识 | 状态所有者、对账与受控审计 | 为资金关联保存，访问可审计 | 非公开标识；不作高基数指标标签 |
| 顾客联系方式、设备和风险原始特征 | 各自业务或风险所有者 | 按目的隔离 | 默认不复制，使用最小风险决策 |
| 配置/模型/策略摘要 | 控制平面、路由服务、审计 | 可持久化版本证据 | 摘要不能夹带秘密或原始训练样本 |

表中的“允许”不是要求所有服务都存一份。每次新增字段，都应记录业务目的、权威来源、允许接收者、存储位置、保留期限、删除路径和批准人。没有这些答案时，默认不把字段加入跨服务契约。

## B.2 从请求链画出信任边界

```mermaid
flowchart LR
    Client["顾客/商户"] --> Hosted["经评估的凭证采集"]
    Hosted --> Vault["密钥库或 PSP 令牌化"]
    Vault -->|"受控引用"| Payment["支付归属者"]
    Payment -->|"最小强类型上下文，不含凭证"| Router["路由服务"]
    Router -->|"决策与精确连接"| Payment
    Payment -->|"固定操作与获授权引用"| Gateway["服务商网关"]
    Secrets["秘密管理器/连接作用域"] --> Gateway
    Gateway --> PSP["原服务商作用域"]
    PSP -->|"已认证回调"| Gateway
    Gateway -->|"已清理、已关联的观察"| Payment
    Payment -->|"允许列表投影"| Audit["决策/审计消费者"]
    Router -.->|"没有直接资金权限"| PSP
```

这是一张目标信任图，不是当前已经部署的密钥库、秘密管理器或网络隔离证明。每一条箭头都应有调用身份、资源范围和数据契约。商户 API 密钥只证明商户入口身份，不能继续转发给路由服务；路由服务获准返回候选通道，不代表它拥有支付网关的执行权限；支付网关能查某个服务商账户，不代表能任意把该账户的事件绑定到另一商户。

同一 VPC、Kubernetes 命名空间或机器不是身份凭证。NIST 的零信任模型明确反对仅因物理或网络位置而隐式信任资源；这里采用的是这条原则，而不是要求为每个示例部署一整套服务网格。[NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final)

生产方案应给支付、路由、网关、配置和结算服务分配独立服务身份。授权应覆盖“谁能对哪个商户、哪条连接、哪种操作做什么”，而不只是拥有一个共享内部令牌。数据库同理：迁移角色与运行角色分离，运行角色不能执行 DDL，路由服务没有支付表权限；只读审计身份也不能顺手获得所有令牌列。当前 Compose 的共享开发账号不能作为生产权限模型照搬。

## B.3 API 契约应拒绝多余能力，而不是只忽略字段

`map[string]any` 作为上下文很方便，也最容易让安全边界逐渐失效。一个 `debug_payload` 字段就能绕过原本精心设计的七个安全属性；一个忽略未知字段的解码器，又会让调用方误以为其中的覆盖值、令牌或规则已被服务接受。

因此入口应同时限制内容类型、总字节数、嵌套深度、字段集合、类型、版本和业务值；拒绝重复键、非预期大小写变体及尾随第二份 JSON。拒绝响应只返回稳定错误码，不回显原始消息体或错误字段的值。Go 的 `encoding/json` 传统解码行为允许重复对象名，结构体匹配也可能不区分大小写；`DisallowUnknownFields` 不是这些检查的总开关。[Go：encoding/json 行为说明](https://pkg.go.dev/encoding/json)

下面用一个刻意很小的审计投影展示正向允许列表。它不是第 4 章的完整上下文模式：只接受精确字段名、一个嵌套执行对象和短字符串；没有自由扩展字段，重复键、数组、空值和任何未知键都会被拒绝。由于允许的形状固定，恶意嵌套不会触发无界递归。

```go
package auditboundary

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

var ErrBoundary = errors.New("audit projection violates its schema")

type shape map[string]shape

func ValidateAuditJSON(raw []byte) error {
	if len(raw) == 0 || len(raw) > 4096 || !utf8.Valid(raw) {
		return ErrBoundary
	}
	schema := shape{
		"decision_id": nil, "payment_intent_id": nil,
		"policy_version": nil, "outcome": nil,
		"execution": {
			"connection_id": nil, "revision": nil,
			"namespace_id": nil, "provider": nil,
		},
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if err := object(d, schema); err != nil {
		return ErrBoundary
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return ErrBoundary
	}
	return nil
}

func object(d *json.Decoder, schema shape) error {
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return ErrBoundary
	}
	seen := make(map[string]bool, len(schema))
	for d.More() {
		token, err := d.Token()
		key, isString := token.(string)
		if err != nil || !isString || seen[key] {
			return ErrBoundary
		}
		child, allowed := schema[key]
		if !allowed {
			return ErrBoundary
		}
		seen[key] = true
		if child != nil {
			if err := object(d, child); err != nil {
				return ErrBoundary
			}
			continue
		}
		value, err := d.Token()
		text, isString := value.(string)
		if err != nil || !isString || len(text) == 0 || len(text) > 128 {
			return ErrBoundary
		}
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') || len(seen) != len(schema) {
		return ErrBoundary
	}
	return nil
}
```

它通过“不允许额外字段”同时拒绝根部 PAN、嵌套 CVV 和变换大小写的秘密字段，不靠一张永远列不全的敏感词黑名单。下面的反例测试还检查错误文本不会回显测试标记。

```go
package auditboundary

import (
	"strings"
	"testing"
)

func TestAuditBoundary(t *testing.T) {
	const valid = `{"decision_id":"dec_1","payment_intent_id":"pi_1",` +
		`"policy_version":"policy_1","outcome":"captured","execution":{` +
		`"connection_id":"conn_a","revision":"rev_1",` +
		`"namespace_id":"ns_a","provider":"provider_a"}}`
	if err := ValidateAuditJSON([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"root credential": strings.Replace(valid, `"decision_id":`, `"PAN":"test_secret","decision_id":`, 1),
		"nested secret":   strings.Replace(valid, `"revision":`, `"CVV":"test_secret","revision":`, 1),
		"duplicate":       strings.Replace(valid, `"outcome":`, `"outcome":"failed","outcome":`, 1),
		"case variant":    strings.Replace(valid, `"decision_id":`, `"Decision_ID":`, 1),
		"trailing value":  valid + `{}`,
		"array injection": strings.Replace(valid, `"rev_1"`, `["rev_1"]`, 1),
		"null value":      strings.Replace(valid, `"rev_1"`, `null`, 1),
		"oversized":       strings.Repeat(" ", 4097),
		"invalid UTF-8":   valid + string([]byte{0xff}),
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateAuditJSON([]byte(raw))
			if err == nil || strings.Contains(err.Error(), "test_secret") {
				t.Fatalf("unsafe acceptance or error response: %v", err)
			}
		})
	}
}
```

这仍不是通用数据防泄漏（DLP）。若开发者把令牌填进合法的 `decision_id` 值，模式并不知道它的真实含义。生产代码还要从权威对象构造强类型投影，验证标识与结果的业务语义，禁止原请求、动态映射和任意错误对象直接进入日志。正向契约解决“哪些槽位可以存在”，数据来源与访问控制解决“谁能往槽位里放什么”。错误路径、跟踪附带信息、死信队列、备份和崩溃转储都必须纳入同一检查。

## B.4 回调：验签、关联和去重是三件事

有效签名只证明消息符合某个密钥的认证契约，不证明它属于当前操作，也不证明内容第一次出现。处理顺序应明确分层：按端点和连接找到允许的认证材料；依服务商规范验证原文或指定规范化内容；校验账户、对象和操作关联；在持久收件箱中去重或隔离；再提交领域结果。

Stripe 要求签名验证使用未被改写的原始消息体；Adyen 标准回调则使用其规定的字段构造签名内容，部分其他回调类型验证原始消息体。不能写一个“对所有 PSP 都先把 JSON 重新序列化再做 HMAC”的通用算法。[Stripe：签名排查](https://docs.stripe.com/webhooks/signature)、[Adyen：HMAC 验证](https://docs.adyen.com/development-resources/webhooks/secure-webhooks/verify-hmac-signatures/)

```mermaid
sequenceDiagram
    participant P as 服务商
    participant G as 网关入口
    participant I as 持久收件箱
    participant M as 支付领域
    P->>G: 带服务商特定签名的回调
    G->>G: 认证获准端点并解析有界载荷
    G->>G: 匹配固定连接、命名空间和操作
    G->>I: 写入事件身份与已清理证据
    alt 同一事件且含义相同
        I-->>G: 重复，不产生额外状态副作用
    else 同一身份但证据改变
        I-->>G: 隔离冲突
    else 新的有效事件
        I-->>G: 持久结果与发件箱记录
        G->>M: 至少一次投递的规范化观察
        M->>M: 单调归约器与独立收件箱
    end
    G-->>P: 仅在持久接受后确认
```

新事件和重复投递还需要区分。事件标识的持久去重防止重复生效；签名中的时间窗口用于限制捕获消息的恶意重放，两者不能相互替代。Stripe 的签名时间戳属于签名内容，而且每次重新投递会产生新时间戳；应依其契约检查投递时效，不把事件发生时间当成固定五分钟后永远丢弃的理由。[Stripe：回调重放防护](https://docs.stripe.com/webhooks#preventing-replay-attacks)

当前网关模拟器协议是原始消息体的 HMAC-SHA256 加持久事件去重，没有真实 PSP 的时间戳与签名版本协议，不能据此声称已完成完整的防重放保护。它已经区分连接认证权限与命名空间冲突域：同一命名空间的另一连接不自动获得写权限，跨路径关联必须是显式授权的恢复别名。这个约束见配套实现的[不可变服务商连接架构决策](https://github.com/YouEclipse/payments/blob/bc4afde24da708d22f67da72a90a1557b1477173/docs/adr/0002-immutable-provider-connections.md)。

## B.5 密钥轮换不能改变旧操作的执行含义

把秘密移到环境变量，只是避免写进源码，不是完成密钥管理。生产设计应使用可审计的秘密管理器、KMS 或适当的硬件边界，明确生成、分发、使用、轮换、撤销、恢复与销毁职责。应用只拿到必要的使用权限，控制平面和审计记录保存凭证版本引用，不复制明文。[NIST SP 800-57 第 1 部分](https://csrc.nist.gov/pubs/sp/800/57/pt1/r5/final)

轮换时至少区分两件事：验证旧回调的认证材料，以及向原服务商账户发起恢复请求的执行权限。旧秘密仍可验证迟到事件，不意味着它还能创建新交易；新秘密生效，也不意味着旧的状态未知操作可以改用另一个账户。每次变化要有激活窗口、可回滚版本、在途操作清单和撤销证据。

已验证基线的连接注册表刻意拒绝在同一标识与修订版下改变端点、命名空间或回调秘密摘要。当前配置的每个连接标识只解析一个修订版，因此变更使用新的标识与修订版，并保留旧映射服务在途恢复。它尚未实现同一端点多个签名密钥的重叠轮换协议。Adyen 官方要求更换 HMAC 密钥时考虑传播时间并暂时接受前一个密钥；真实适配器应显式实现这种重叠契约，不能通过关闭注册表校验来解决。[Adyen：HMAC 密钥轮换](https://docs.adyen.com/development-resources/webhooks/secure-webhooks/verify-hmac-signatures/#changing-your-hmac-key)

配置平台也是安全边界。能把路由指向攻击者端点的发布权限，可能比读取一条日志更危险。作者、审批者、发布者应分权；制品必须带版本与摘要，完整校验后原子激活；运行实例不能接受请求中临时注入的规则或端点。紧急破障需要短时权限、具体理由、审批或事后复核及审计记录，不能拥有“把状态未知一键改失败”的无条件权限。

## B.6 保留、删除与跨区处理要设计在模式里

仅追加说明事实不能被任意改写，不等于把所有个人数据永久保存。[第 14 章](../part-4-operating-at-scale/routing-decision-store.md)的证据记录应尽量引用版本与摘要，把需要短期保留的敏感映射分离；删除、匿名化或限制处理必须有可验证流程，而非口头承诺“日志过几天自动清理”。

以 GDPR 适用场景为例，删除请求存在需要评估的条件，也存在法定义务或法律请求等保留例外；不能为了方便对账就一律拒绝删除，也不能收到请求便删除必须保留的资金事实。由合规与法律负责人确定目的和依据，再由工程实现对应的数据类、保留策略与访问限制。[欧盟委员会：处理个人权利请求](https://commission.europa.eu/law/law-topic/data-protection/information-business-and-organisations/dealing-requests-individuals_en)

工程上至少维护一份数据去向清单：在线表、只读副本、分析仓、对象存储、日志、跟踪数据、死信、导出文件和备份。备份的保留周期、恢复后重新执行删除标记的程序、诉讼保留与正常 TTL 的优先级都要明确。哈希也不天然匿名；低熵值可被枚举，与别表重新关联也可能恢复身份，因此“只存摘要”不能自动结束隐私评估。

多区域的数据驻留同样不是一个 `country=DE` 的布尔条件。应区分交易处理位置、持久存储位置、支持人员访问位置和跨区分析用途。路由服务只消费经过批准的合格范围与版本，不自行解释全部法条；不确定的授权范围必须拒绝。区域故障时，业务方希望继续收款，也不能因此默许把凭证或敏感证据复制到未获批准的区域。[第 19 章](../part-5-global-routing/multi-region.md)讨论的容灾权限必须包含这道门。

## B.7 把威胁转成可以失败的测试

| 威胁 | 必须观察到的失败或隔离 | 证明材料 |
| --- | --- | --- |
| PAN/CVV/秘密被塞进根部、嵌套或大小写变体字段 | 请求被拒绝，错误和日志不回显值 | HTTP/契约测试、日志捕获与夹具扫描 |
| 同一事件 ID 改写状态或关联 | 收件箱冲突被持久隔离，不覆盖资金事实 | 支付网关与支付服务双侧数据库测试 |
| A 的合法签名用于更新 B 的操作 | 未授权关联失败，即使命名空间相同 | 不同连接与别名的正反例 |
| 同一修订版重绑端点或密钥 | 启动或就绪检查失败，不继续原未知操作 | 注册表重启与漂移测试 |
| 路由身份尝试执行或读取支付表 | 授权和数据库权限拒绝 | 真实部署身份/角色的负向测试 |
| 生产配置使用模拟器或开发凭证 | 启动或上线门禁失败 | 环境矩阵及真实启动结果 |
| 调试日志、程序崩溃、跟踪数据或死信泄漏受控引用 | 测试捕获到泄漏即失败 | 合成敏感标记，不使用真实卡数据 |
| 密钥撤销、备份恢复或跨区切换绕过边界 | 未授权使用被拒绝，在途恢复按原契约处理 | 演练记录与审计关联 ID |

不要只搜索源码中有没有 `cvv`。泄漏可能来自 `%+v` 打印整个结构体、HTTP 中间件记录请求头、SDK 自动追踪消息体，或异常把 URL 查询参数写入日志。测试应使用合成标记走成功、拒绝、超时、解析失败和程序崩溃恢复路径，再检查所有输出面。敏感值不能作为指标标签；普通支付标识也会造成高基数和关联风险，应按[第 15 章](../part-4-operating-at-scale/observability.md)的有限维度观察。

## B.8 当前实现与上线剩余工作

当前固定基线已经提供商户身份绑定、支付/路由/网关内部令牌校验、严格且有限体积的 JSON、模拟器环境限制、精确连接关联、HMAC 回调、持久收件箱与隔离区、不可变修订版注册表，以及禁止 `UPDATE`、`DELETE`、`TRUNCATE` 的服务商观察账本；这些边界对应[固定代码版本](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)及其[持续集成记录](https://github.com/YouEclipse/payments/actions/runs/33041147119)。路由服务的共享令牌中间件保护除 `/livez`、`/readyz` 和 `/metrics` 外的所有路径，检查 `X-Internal-Token`，等长值使用恒时比较，无令牌或错误令牌返回 401；开发环境允许本地默认值，非开发环境未显式配置则拒绝启动，入口位于 `services/router/cmd/server/main.go`。

但它仍是本地与测试支付核心。已有共享令牌鉴权不等于已具备 mTLS、每服务独立主体或细粒度资源授权；Compose 使用共享开发数据库账号；全链路日志脱敏、秘密管理器/KMS、独立生产数据库角色、真实 PSP 时间戳/重放防护与重叠轮换、数据保留与删除执行器、合规评估，均不能从现有通过的 CI 推导为已经完成。上面的 Go 允许列表是可运行的附录参考，也不等于已经挂到所有生产输出入口。书稿和实现仓库当前均为私有，证据链接需要获授权的 GitHub 身份；这是一项访问边界，不是修改仓库可见性的理由。

这些缺口应进入[附录 C 的上线门禁](capacity-planning-and-launch-checklist.md)，由明确负责人提供部署证据。安全边界不是写到文档末尾的免责条款：只有错误身份、错误字段和错误配置真的被系统拒绝，正文中画出的边界才具有运行意义。
