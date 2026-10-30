# Designing Modern Payment Systems

[中文](../README.md) | **English**

## From a Single Payment to a Global Payment Routing Platform

The central challenge in a modern payment system is more than calling a payment service provider's API. It is making safe, explainable, and operable payment decisions across countries, currencies, payment methods, merchant entities, and providers.

This book follows the evolution of a **payment routing system**, starting with the simplest architecture: a single provider. As the business encounters differences in coverage, authorization rates, processing costs, risk constraints, provider failures, and global deployment, the design develops routing context, a rule engine, eligibility filtering, dynamic scoring, fallback routing, decision auditing, and reconciliation.

This is not an encyclopedia of payment terminology, nor does it begin with a complex architecture and explain it backward. The major chapters follow this progression:

```text
Problem → Simple Design → Problems at Scale → Production Design → Trade-offs
```

You first see how a sufficiently simple system works, then the real conditions under which it fails, and finally why new components are needed, what they solve, and what they cost.

> This edition contains 22 chapters, three appendices, and a glossary. Distinguish capabilities already implemented in the companion system from Go teaching models and target designs still awaiting implementation. Each chapter states its scope and limitations.

## Language Conventions

The Chinese and English editions follow the same chapter structure and payment semantics. Use the language link at the top of a page to switch to its counterpart. Brand names, standards abbreviations, code identifiers, API fields, state enums, file paths, and test names retain their original spelling. The [glossary](glossary.md) provides a consistent vocabulary for the book. Original hand-drawn illustrations retain their Chinese labels, with English captions explaining their meaning; technical diagrams are translated into English.

## Code Example Conventions

All backend implementation examples use **Go**, including domain models, APIs, payment state machines, routing-rule evaluation, provider adapters, idempotency and retries, data access, and concurrency. Examples favor clear types and the standard library and aim to be compilable and testable rather than merely illustrative fragments.

JSON, SQL, Mermaid, and occasional pseudocode express API data, storage models, architecture, interactions, and decision formulas. They serve those specific explanatory purposes rather than introducing a second backend implementation language.

## The Companion Go Payment System

The book maintains a runnable microservice implementation: [YouEclipse/payments](https://github.com/YouEclipse/payments). It is more than a collection of chapter snippets. A payment service, routing service, provider gateway, and provider simulator restricted to development and testing form a vertical flow backed by real PostgreSQL.

The current implementation focuses on automatic capture for tokenized card payments and treats timeouts as `UNKNOWN`. It queries the same provider operation or waits for a webhook rather than blindly switching PSPs when a response is missing. Router v3 makes constrained, explainable decisions; the payment service retains funds state and recovery budgets; and the gateway retains immutable connection identities and append-only external evidence. The simulator does not handle real money. Production PSP integration still requires real connectors, key management, security and compliance work, and operational readiness. The result is a runnable, verifiable payment core rather than a finished platform ready to process live funds.

This edition pins the companion implementation to [bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173), with [successful CI run 33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119). Chapters and appendices distinguish the current implementation, teaching models, and future designs. Passing tests for the companion system does not imply that every capability discussed in the book has been implemented.

## Who This Book Is For

The primary audience includes:

- Backend engineers with roughly three to eight years of experience;
- Payment and fintech engineers;
- System architects and infrastructure engineers;
- Engineers entering the payments domain;
- Technically experienced payment product managers who want to understand system constraints.

The book assumes familiarity with HTTP, REST APIs, databases, Redis, message queues, microservices, and basic consistency and availability concepts in distributed systems. It explains payment terminology and funds semantics without reteaching general backend engineering.

## What You Will Learn

After reading, you should be able to:

- Explain how customers, merchants, PSPs, acquirers, card networks, and issuers complete a card payment;
- Define the responsibilities of payment APIs, orchestration, routing, and provider adapters;
- Build routing context and exclude ineligible paths before selection;
- Design versioned, verifiable routing rules with gradual rollout and rollback;
- Evaluate provider performance under sample-size limits, data delay, and selection bias;
- Combine success rate, cost, latency, risk, and health into explainable decisions;
- Handle timeouts, unknown outcomes, retries, fallback, and idempotency without duplicate charges;
- Build decision storage, key metrics, experiments, and audit capabilities;
- Converge platform and provider state through webhooks, queries, and reconciliation;
- Design a routing platform spanning countries, currencies, payment methods, PSPs, and regions;
- Judge where rules, statistics, contextual bandits, and machine learning fit, rather than adding models merely to make a system appear intelligent.

## Book Structure

### Part I: Payment Foundations

Follow the complete lifecycle of a card payment and establish PSP, acquiring, card-network, issuer, authorization, capture, refund, and chargeback concepts. Then derive payment routing from a single-provider architecture and establish the boundary between routing and orchestration.

### Part II: Building the Routing Engine

Develop routing context, rule models, provider eligibility, cost models, success-rate estimates, risk constraints, and provider scoring. Use a common adapter boundary to isolate different PSP interfaces and state semantics.

### Part III: Reliability

Address one of payments' most dangerous failures: the request times out, but the funds operation may already have succeeded. Define payment, attempt, and operation idempotency, distinguish safe and unsafe fallback, and design provider health and circuit breakers.

### Part IV: Operating Payment Routing at Scale

Move from making decisions to operating them over time. Record decision evidence, establish observability and SLOs, run controlled experiments, repair discrepancies through reconciliation, and connect decisions to actual fees, settlement, and chargebacks.

### Part V: Global Payments and Routing Evolution

Explore multi-region deployment, data residency, configuration releases, security and compliance boundaries, and when statistical routing, contextual bandits, or machine learning should supplement rules.

### Part VI: An Integrated Case Study

Design a global routing platform targeting 10,000 merchants, 50 countries, 30 currencies, 20 payment methods, 15 PSPs, and a peak of 100,000 TPS. Connect APIs, state machines, storage, caching, messaging, decisions, failover, reconciliation, and deployment.

The appendices provide state models and error classifications for payment intents, attempts, and provider requests; security, compliance, and data boundaries; and a capacity-planning and launch checklist. The [glossary](glossary.md) standardizes terminology throughout both editions.

## Where to Start

- New to payments: follow the [table of contents](SUMMARY.md) in order.
- Experienced with payment integrations: start with [Why Payment Routing?](part-1-foundations/why-payment-routing.md), then continue to the routing-engine and reliability sections.
- Designing a multi-PSP platform: focus on system boundaries, eligibility, idempotency, fallback, decision auditing, and reconciliation.

## Scope and Writing Principles

### Working Subtitle

Payment Routing, Reliability, Cost Optimization, and Risk Control at Scale

### Book Positioning

Designing Modern Payment Systems is a practical system design book about building modern payment infrastructure, with payment routing as the central system. The book explains how payment platforms route transactions across PSPs, acquirers, payment methods, regions, currencies, cost models, risk signals, and reliability constraints.

### Target Readers

The primary readers are backend engineers, payment engineers, fintech architects, infrastructure engineers, and technical product managers who want to understand how modern payment routing systems are designed.

The reader is expected to understand basic backend systems, APIs, databases, queues, and distributed systems concepts. Prior payment experience is helpful but not required.

### Core Promise

After reading this book, the reader should be able to design a payment routing system that is reliable, observable, cost-aware, risk-aware, and scalable across multiple payment providers and markets.

### Core System

The central system in this book is the Payment Routing System. It receives a payment request and decides where and how the transaction should be processed.

A routing decision may consider provider availability, payment method, card brand, issuer country, merchant configuration, transaction amount, currency, cost, historical success rate, risk signals, compliance constraints, latency, and fallback options.

### Scope

This book covers payment routing, PSP integration, payment lifecycle, rule-based routing, cost-aware routing, risk-based routing, retries, fallbacks, observability, reconciliation, settlement implications, experimentation, and global payment infrastructure.

### Non-Goals

This book is not a general introduction to consumer payments, a legal compliance manual, a PCI implementation checklist, or a complete accounting textbook. Topics such as PCI, ledger design, settlement, and reconciliation are covered only to the extent that they influence payment routing system design.

### Writing Principles

Each chapter should connect domain concepts with system design decisions.

Each major concept should answer four questions:

1. What problem does this solve?
2. What can go wrong?
3. What does a simple design look like?
4. How does the design evolve at scale?

The book should prefer concrete architecture, data models, state machines, failure modes, and operational trade-offs over abstract explanations.
