# Glossary

[中文](../glossary.md) | **English**

This glossary follows the book's central theme of payment routing. Definitions emphasize each term's meaning in architecture and routing decisions. Specific rules differ across countries, card networks, and payment providers; implementations must follow the relevant contracts, API documentation, and applicable regulations. The Chinese edition includes English equivalents to connect its terminology with external specifications and code.

## Numbers and Symbols

### 3-D Secure (3DS)

A cardholder authentication protocol for online card payments. A provider's support for 3DS versions, frictionless and challenge flows, and the transmission of strong customer authentication exemption information can inform eligibility filtering or risk-aware routing. A frictionless flow is still an authentication flow; it is not equivalent to an exemption from strong customer authentication.

### Strong Customer Authentication (SCA)

An authentication mechanism that requires multiple independent factors to verify the payer, particularly common in European payment contexts. The routing service generally consumes authentication requirements supplied upstream and selects a route that can perform the required flow. It does not determine legal applicability itself.

## A

### Acquirer

A licensed institution that provides card acquiring services to merchants and submits transactions to card networks. A PSP may act as the acquirer directly or connect to one or more acquirers.

### Active-Active

A deployment model in which multiple regions serve production traffic simultaneously. For payment routing, it also requires handling payment-intent ownership, global idempotency, configuration versions, and cross-region fallback. Replicating stateless services alone is insufficient.

### Actual Cost

Cost derived from provider settlement rows, fee details, or confirmed economic facts. Its source, currency, matching scope, and revision relationships must be retained. It must not overwrite the expected cost recorded at decision time. Unreceived or incompletely matched amounts must not be presented as final costs.

### Authorization

The process of approving or declining a payment request, usually decided by the issuer. Under specific conditions, the network or processor may act on the issuer's behalf according to established rules. Approval normally means that credit or funds have been reserved; it does not mean that the merchant has received the money or that subsequent settlement is guaranteed.

### Authorization Rate

The proportion of authorizations approved. Whether the numerator and denominator include technical failures, repeated attempts, or risk blocks must be explicit. For routing, the metric must also identify the provider, country, issuer, payment method, time window, and sample size.

### Address Verification Service (AVS)

A capability that compares an address supplied by the payer against issuer records. AVS coverage and result semantics vary by market and provider. It can inform risk decisions or route capabilities, but it is not a globally uniform Boolean field.

## B

### Bank Identification Number / Issuer Identification Number (BIN / IIN)

The card-number prefix used to identify an issuing institution, card product, and related attributes. BIN and IIN can help segment provider performance, but they are payment data requiring careful handling. Their lengths and metadata must not be hard-coded as universal assumptions.

### Bulkhead

Isolation of connection pools, threads, concurrency allowances, or queues by provider or another failure domain. It prevents one provider's latency from exhausting the resources of the entire routing service.

## C

### Cardholder Data (CHD)

Protected account data as defined by PCI DSS, including at least the full PAN. When present with the PAN, it may also include the cardholder's name, expiration date, and service code. CHD may be stored where a business need exists and applicable controls are satisfied; it must not be confused with SAD, which cannot be retained after authorization.

### Card Verification Value (CVV / CVC / CID)

A security code used in card-not-present and similar transactions to help verify that the cardholder possesses the card. It is SAD and must not be retained after authorization, even in encrypted form. The routing service must not write it into context, decisions, logs, traces, or messages.

### Route Candidate

A specific route capable of executing the current payment attempt, usually identified by a combination of provider, merchant account, connector, and region. Scoring compares route candidates, not merely abstract PSP brand names.

### Capability Catalog

Versioned data describing the countries, currencies, payment methods, capture modes, authentication methods, and other capabilities supported by each route candidate. It is an important input to eligibility filtering, but it is distinct from dynamic provider health.

### Connection Revision

An immutable reference to an execution connection configuration. The payment attempt, funds operation, and provider gateway pin the same identity and revision. Configuration rotation creates a new revision; an old revision cannot be rebound to a different account or endpoint. An existing `UNKNOWN` operation does not migrate with a new configuration.

### Config Bundle

A release unit that pins references and digests for immutable artifacts such as rules, capabilities, fees, experiments, and models. Validation, approval, and runtime loading must all target the same bundle. Runtime activation is atomic; individual pieces must not be replaced piecemeal.

### Capture

An operation following authorization that confirms all or part of the authorized amount for clearing and settlement. Differences in support for automatic and manual capture affect provider eligibility.

### Card Network / Card Scheme

A network, such as Visa or Mastercard, that connects acquirers and issuers, defines card-transaction rules, and carries authorization and clearing information. Its fees and rules can indirectly affect routing costs through provider contracts.

### Chargeback

A mechanism for recovering funds from a merchant through card-network procedures, triggered by a cardholder dispute or another qualifying reason. A chargeback can occur weeks or months after payment, creating delayed feedback for risk-adjusted costs.

### Circuit Breaker

A mechanism that temporarily blocks new requests after provider failures cross a threshold, then uses limited probes during recovery to determine whether traffic can resume. Circuit state should follow actual failure dimensions rather than always acting as a global switch.

### Clearing

The process by which participants exchange, validate, and aggregate transaction details and calculate amounts receivable and payable. It normally occurs after authorization and before settlement.

### Connector

An implementation of a provider protocol, API version, field mapping, and state mapping. It is not a complete provider connection specifying an account, region, endpoint, and credential scope. One connector implementation may serve multiple exact connections, but it cannot independently replace the connection or reroute an operation.

### Control Plane

The system that manages creation, validation, approval, and release of rules, costs, capabilities, models, and experiments. It can be separated from the data plane that serves real-time routing decisions.

### Cross-Border Fee

An additional fee that may arise when transaction participants are in different countries or regions. The determination does not necessarily depend only on the customer's country; it must reflect the actual rules of the provider and card network.

## D

### Data Plane

The components that build context, filter eligible candidates, apply rules and scoring, and return routing decisions in the online payment path. They usually operate under strict latency and availability budgets.

### Data Residency

A constraint requiring particular categories of data to be stored or processed in specified jurisdictions. It affects region selection, context fields, decision logs, and offline metric aggregation.

### Decline

A business outcome in which authorization is not approved. Issuer or network declines must be attributed separately from provider risk blocks before authorization, gateway rejections, and integration errors. Adapters normalize decline reason codes while preserving the original provider codes. Whether a soft or hard decline is retryable depends on the payment method, network guidance, and provider semantics.

### Decision Reason Code

A structured explanation of why a candidate was excluded, had its score adjusted, or ultimately won. It is better suited than free-form logs to auditing, metric aggregation, and stable API contracts.

### Digest

A content fingerprint computed over explicitly canonicalized context, artifacts, or decisions. A digest can check content consistency, but it does not by itself establish publisher identity, approval authority, or factual correctness. Reusing a version name must not conceal different content.

## E

### Provider Eligibility

Filtering all provider routes down to those capable of handling the current transaction, using hard constraints such as capabilities, contracts, compliance, and unavailability. `Unavailable` or `CircuitOpen` may cause hard exclusion; `Degraded` usually leads to a scoring penalty or traffic limit. Eligibility filtering runs before preference scoring.

### Effective Sample Size (ESS)

A diagnostic of how much information weighted samples effectively contain, often calculated as `(Σw)² / Σw²`. A few very large weights can leave a large dataset with a small ESS. ESS cannot repair missing policy support, incorrect action probabilities, or immature outcomes.

### Experiment Exposure

The event in which a business unit actually receives an experimental treatment. Assignment only establishes a bucket; it does not mean that a candidate was executed. A shadow decision does not create funds exposure either. Metrics must retain assignment, decision, execution, and outcome separately, avoiding sample selection based only on requests that were successfully delivered.

### Expected Cost

An estimate made before the payment outcome is final, incorporating rates, success probabilities, foreign exchange, failed attempts, and potential risk. It must be distinguished from actual cost obtained after settlement and continuously calibrated against it.

## F

### Fallback Routing

A recovery process that selects another eligible route and creates a new payment attempt after the preceding attempt reaches a definite outcome that permits rerouting. Retrying the same funds-operation identity with the same provider is not fallback. An unknown outcome does not permit an immediate switch to another provider.

### Fencing Token / Ownership Epoch

A monotonically increasing generation associated with transfers of execution ownership. Write transactions and side-effect entry points validate the current ownership generation and reject stale owners. DNS, lease expiry, and worker liveness cannot replace this check. Fencing cannot retract a request already sent to a provider; the original funds operation must still be recovered.

### Foreign Exchange (FX)

Conversion between currencies, including the exchange rate and markup. Which party performs the conversion and which rate is used for settlement affect both the true routing cost and the amount the merchant receives.

### Fraud Engine

A system that evaluates transaction fraud risk and produces a decision, score, or reasons. A payment routing service may consume its results, but normally should not replicate model features or take over responsibility for fraud decisions.

## G

### Payment Gateway

A system that receives merchant payment requests, provides secure access or protocol translation, and connects to downstream processors. The roles of a gateway, PSP, payment orchestrator, and routing service may overlap in commercial products. This book distinguishes them by architectural responsibility rather than brand name.

### Guardrail

A constraint that optimization or experimentation must not violate, such as compliance eligibility, a risk ceiling, provider capacity, or a maximum number of fallback attempts. A higher model score cannot compensate for violating a guardrail.

## H

### Hard Decline

A business decline that usually indicates the request should not be retried unchanged, such as a credential reported as invalid or prohibited. Classification is not fully consistent across providers. The routing system determines the action from normalized codes and an explicit policy.

### Provider Health

Operational status derived from latency, timeouts, technical errors, active probes, and similar signals. It must carry dimensions, timestamps, and confidence; it is not interchangeable with authorization rate.

### Home Region

The region responsible for a payment intent's authoritative state and execution ownership. It differs from the ingress region receiving the HTTP request, the execution region in which a worker runs, and the external provider's region. Moving traffic does not automatically transfer authority to write funds-related state.

## I

### Idempotency

The property that submitting the same business operation repeatedly does not create additional unintended side effects. A provider may return the original response, the current object state, or a conflict error, so response semantics need not be byte-for-byte identical. A reliable implementation requires business identity, durable state, and concurrency constraints, not just a response cache.

### Idempotency Key

A stable identifier supplied by the caller for one logical operation. The system must define its scope, retention period, whether parameters can change, and the error semantics for the same key with a different request.

### Integrity Status

An independent dimension describing whether evidence within an aggregate can safely be interpreted automatically. This book uses `healthy` and `review_required`. Requiring manual review does not erase an existing capture fact or turn `succeeded` into `failed`.

### Inbox / Outbox

Durable patterns for recording received message identities and events awaiting publication within a transaction, respectively. They commit with domain state in the same transaction to address process-crash windows. Delivery can still be duplicated; downstream consumers must still validate idempotency using event IDs and content. These patterns do not establish end-to-end exactly-once delivery.

### Issuer

An institution that issues cards or payment accounts to customers and approves or declines transactions during authorization based on account status, funds, risk, and rules.

## L

### Ledger

A system that records changes in financial rights and obligations as auditable entries. The routing service may reference payment and settlement outcomes, but routing decision storage must not be treated as a funds ledger.

### Last-Known-Good (LKG)

The most recent configuration that passed complete validation and loaded successfully. It may remain in use during a temporary control-plane outage, subject to its validity period and safety boundaries. An expired configuration, a mandatory disablement, or missing required artifacts cannot be used indefinitely under the label of LKG.

## M

### Merchant

The entity that supplies goods or services to customers and initiates collection. Its contracts, business category, country, and policy scope influence route eligibility and routing policy.

### Merchant Account

The account or contractual entity through which a merchant or payment platform processes transactions with a particular provider. Rates, currency capabilities, limits, and risk policies are usually bound to this account, not merely to the provider's brand name.

### Payment Method

The instrument or network a customer uses to pay, such as a card, digital wallet, bank transfer, or local payment method. Differences in synchrony, retryability, authentication, and finality directly constrain routing.

## O

### Off-Policy Evaluation (OPE)

Estimating the outcomes of a target policy using data generated by a historical behavior policy. Rewards, maturity windows, logged action probabilities, and support sets must be explicit. An action that the target policy would choose but that had zero historical probability cannot be identified through reweighting. OPE results are not a promise of causal improvement without validation in a controlled online experiment.

### Payment Orchestrator

The component that owns the payment-intent lifecycle and coordinates risk systems, routing, provider adapters, asynchronous events, and recovery. The routing service produces an immutable route plan and performs guardrail-only revalidation. The payment orchestrator decides whether evidence from the previous attempt permits a new funds path; an execution binding must not masquerade as a new routing ranking.

## P

### Primary Account Number (PAN)

The number identifying a card account. PAN is CHD, not SAD. This book's routing context receives only necessary derived attributes and contains no PAN, CVV, or token that can execute a payment. Tokens flow only within the controlled payment-service and gateway boundary; being different from the original card number does not justify including them in logs or decisions. [PCI SSC Glossary](https://www.pcisecuritystandards.org/glossary/)

### Payment Attempt

One controlled execution toward the same payment intent, pinning the payment method and provider connection. A network retry of the same logical funds operation does not create a new attempt. Only a policy-permitted fallback after the previous attempt is definitively known to have no side effects can create another attempt. `UNKNOWN` blocks new funds paths.

### Payment Intent

A merchant's business objective to collect a particular amount in a particular currency. It may pass through authentication, multiple attempts, capture, or recovery from failure. Its lifecycle is generally owned by the payment orchestrator.

### Payment Service Provider (PSP)

An organization providing merchants or platforms with one or more services such as payment access, processing, acquiring, or aggregation of payment methods. PSP is a broad commercial term; its actual capabilities must be modeled through capability descriptions and connectors.

### Payment Processor

An organization or system providing transaction-message processing, connectivity, or back-office processing. It may be the same entity as the PSP or acquirer, or merely a technical service provider in the chain.

### Tokenization

A technique that replaces sensitive credentials such as PAN with controlled tokens. A token's usability may be limited by provider, merchant, network, or region. This portability affects fallback routing.

### Provider Adapter

An adaptation layer that maps the platform's common payment operations to a provider's API and normalizes responses, error categories, and webhook events. It must explicitly expose differences in idempotency scope, retention, concurrency, and response semantics rather than hide them behind a uniform interface. It executes the selected route and must not secretly apply an independent routing policy.

### Provider Connection

An exactly executable configuration of provider, merchant account, connector, and region, including endpoint and credential scope controlled by the provider gateway. The routing service returns a non-secret reference. The payment service pins its identity, revision, and external namespace; the execution layer cannot make a second ad hoc selection.

### Provider Namespace

The external identity collision domain declared by an adapter according to the provider's contract. It cannot be inferred from the provider's brand name alone. If transaction references, events, and idempotency keys have different scopes, those scopes must be modeled separately. Connections in the same namespace do not thereby share webhook authentication authority.

### Provider Operation

One logical funds operation, such as `SALE`, `AUTHORIZE`, `CAPTURE`, `VOID`, or `REFUND`. Its internal operation identity, provider idempotency key, semantic payload, and exact execution connection remain fixed. There may be multiple network requests, but only one funds-operation identity. When the text refers generally to a provider request, distinguish the logical operation from one transmission.

### Propensity and Support

Propensity is the probability that the behavior policy selected an action in the context at that time; it is not the action's predicted success rate. Off-policy evaluation requires the logging policy to cover the target policy's entire action support. Checking only that the observed action had a positive probability is insufficient.

## R

### Reconciliation

Matching the platform's internal payment and provider states against provider APIs, webhooks, transaction files, or settlement files to detect missing or duplicate records, state discrepancies, and amount discrepancies, then trigger controlled repair.

### Refund

A merchant-initiated return of all or part of a captured amount. Refunds differ from voids in timing and network semantics and may incur separate fees and asynchronous states.

### Retry

Another execution of the same logical operation. A payment retry pins the original funds operation, provider, account, region, revision, idempotency key, and semantic payload. It is constrained by outcome certainty, the provider contract, the overall deadline, and the retry budget. It does not create a new payment attempt and is distinct from fallback.

### Risk-Aware Routing

Routing that incorporates risk decisions, authentication requirements, and provider risk-control capabilities into eligibility or scoring. Its purpose is to select an execution path within the risk system's permitted boundaries, not to have the routing service make a new fraud judgment.

### Routing Context

The minimal versioned input pinned for a decision by the payment service's context builder. It includes business fields and necessary non-secret derived attributes. Every observation carries its source, version, time, and a status such as known, unknown, or stale; missing values must not masquerade as default truths. Dynamic health and performance estimates can participate through separate read-only snapshots, but their versions must be recorded. Unaudited live queries must not be mixed into the decision.

### Routing Decision

The routing service's selection of a candidate or an explicit negative outcome, together with reasons and version evidence. A `selected` branch can form an immutable `RoutePlanDecision`. Negative branches such as `no_eligible_candidate`, `blocked`, `unsupported`, `requires_review`, or dependency unavailability must not fabricate a payment attempt, selection, or provider operation.

### RoutePlanDecision

An immutable execution-permission boundary produced by the routing service. It pins candidate order, policy version, capability version, feature version, maximum number of attempts, budgets, and validity period. Guardrail-only revalidation during fallback can only veto the pinned next candidate; it cannot add candidates, reorder, rescore, or extend the plan. Execution stops when the plan expires or evidence is unavailable.

### FallbackExecutionBinding

An immutable object written by the payment orchestrator in an authorized successor transaction. It references the original `RoutePlanDecision`, preceding attempt, pinned next candidate, fresh risk evidence, guardrail evidence, and budget changes. It explains why the payment service created that attempt, but is not a routing decision and does not authorize execution outside the plan.

### Routing Decision Record

An immutable factual record used to explain and reproduce a routing decision. It includes candidates, exclusion reasons, score breakdowns, version references, and latency, but should not duplicate unnecessary sensitive data.

### Routing Policy

Business rules and optimization objectives governing how candidates are filtered, preferred, allocated traffic, or selected. Policy execution belongs to the data plane; creation, approval, and release normally belong to the control plane.

### Routing Rule Engine

A component that executes versioned routing rules with deterministic semantics, handling condition matching, priorities, and conflicts. It is not the complete routing service: it works alongside eligibility filtering, features, scoring, and selection.

### Routing Configuration Platform

A control-plane system for editing, validating, replaying, approving, publishing, gradually rolling out, and rolling back routing policies and related configuration.

### Recovery Point Objective / Recovery Time Objective (RPO / RTO)

The acceptable data-loss window and the target time to restore service, respectively. Payment failover must separately assess the RPO for execution identity, idempotency, and funds evidence. Overall database replication lag does not establish that an unknown funds operation can safely be replayed.

## S

### Sensitive Authentication Data (SAD)

Security data used to authenticate cardholders, including full track data, CVV, CVC, CID, PINs, and PIN blocks. Under PCI DSS, this data must not be retained after authorization, even if encrypted. It must not enter routing context, decision storage, logs, or message queues.

### Settlement

The transfer of net or gross funds between participating institutions based on clearing results. It generally occurs after authorization and capture. Timing and fees depend on the provider, currency, and market rules.

### Scheme Fee

A fee charged by a card network for a transaction or related service. Its structure can be complex and may be passed through an acquirer or PSP. Cost models must be calibrated against actual contracts and settlement data.

### Provider Scoring

Computing relative preferences among candidates that already satisfy hard constraints, based on objectives such as success rate, cost, latency, and risk. Scoring cannot override eligibility or safety guardrails.

### Shadow Decision

A challenger policy's decision computed on the same snapshot as a real request to compare disagreement and operational viability. It does not send a provider request or create a second funds operation, and it does not treat an unexecuted candidate's outcome as observed.

### Side-Effect Certainty

An evidence dimension describing whether an operation is known not to have occurred, may have occurred, or is confirmed to have occurred. It differs from HTTP errors, business success rates, and retryability. `retryable=true` does not authorize switching PSPs. A `not_found` result can establish absence of side effects only under an explicit query contract.

### Settlement Line

An immutable economic fact from a provider report, carrying file identity, row identity, external namespace, transaction reference, amount, currency, and type. Repeated identical content is idempotent; corrections append new facts. A settlement line is neither a new payment-intent state nor a substitute for an accounting entry.

### Soft Decline

A business decline that may succeed after incorrect data is corrected, required authentication is completed, or a later retry is made according to network and issuer guidance. Whether retry or fallback is allowed must follow normalized decline reasons and an explicit policy, not merely an HTTP status code. The original transaction's semantics must not be changed arbitrarily to obtain approval.

### Static Routing

Selecting a route using a fixed provider, fixed priority order, or simple proportions. It is easy to implement and explain, making it useful as a starting point or safe degradation policy, but it cannot adapt to fine-grained performance changes.

### Success-Rate Routing

Estimating each candidate's probability of success from historical performance relevant to the current transaction and using those estimates in selection. A reliable implementation must handle segmentation, sample size, noise, cold starts, and data latency.

## T

### Traffic Splitting

A mechanism that deterministically assigns some transactions to different providers or policies, for contractual quotas, capacity control, or experiments. Experimental allocation should keep the same business unit in a stable group.

### Transaction

A general term for a payment-related business or network record. Different systems use it for payment intents, attempts, authorizations, or settlement lines, so data models and APIs should use more specific object names.

## U

### Unknown State / Unknown Result

A state in which the caller cannot determine whether the provider has produced a side effect, for example when the provider approved the operation but its response was lost. Unknown is not an ordinary failure. Switching providers before a query, webhook, or reconciliation confirms the result can cause duplicate payment.

## V

### Void

An operation that cancels the original transaction before settlement completes, while the provider's state still permits cancellation. Some providers use void specifically to release an uncaptured authorization; others also allow cancellation of captured or submitted transactions that have not settled. Adapters must distinguish authorization reversal, cancellation, and void, and normalize their respective state windows and outcome semantics.

## W

### Webhook

An HTTP callback through which a provider sends an asynchronous state event to the platform. Receivers must verify the source and handle duplicates, reordering, delay, and delivery failure. Webhooks cannot be assumed to arrive exactly once or in strict order.

## Compound Terms Related to Routing Policy

### Cost-Aware Routing

Routing that considers expected costs such as fixed and percentage fees, cross-border fees, foreign exchange, and failed attempts. The lowest quote for one attempt does not necessarily yield the lowest cost per successful payment.

### Intelligent Routing

An umbrella term for dynamically optimizing routing through statistical estimates, contextual bandits, machine learning, or similar methods. Intelligent routing remains constrained by eligibility, compliance, risk, capacity, and reliability guardrails.

### Rule-Based Routing

Routing that expresses deterministic policies for merchants, countries, currencies, and similar dimensions through declarative conditions and actions. Production systems also require rule versions, conflict detection, validation, approval, and rollback.
