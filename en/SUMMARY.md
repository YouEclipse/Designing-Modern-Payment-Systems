# Table of Contents

* [Introduction](README.md)
* [Companion Go Implementation](https://github.com/YouEclipse/payments)

## Part I: Payment Foundations

* [Chapter 1: The Lifecycle of a Payment](part-1-foundations/payment-lifecycle.md)
* [Chapter 2: Why Payment Routing?](part-1-foundations/why-payment-routing.md)
* [Chapter 3: Responsibilities and Boundaries of a Payment Routing System](part-1-foundations/routing-system-boundaries.md)

## Part II: Building the Routing Engine

* [Chapter 4: Building Routing Context](part-2-routing-engine/routing-context.md)
* [Chapter 5: Rule-Based Routing](part-2-routing-engine/rule-based-routing.md)
* [Chapter 6: Provider Eligibility](part-2-routing-engine/provider-eligibility.md)
* [Chapter 7: Cost-Aware Routing](part-2-routing-engine/cost-aware-routing.md)
* [Chapter 8: Success-Rate Routing](part-2-routing-engine/success-rate-routing.md)
* [Chapter 9: Provider Scoring and Risk-Aware Routing](part-2-routing-engine/provider-scoring.md)

## Part III: Reliability

* [Chapter 10: Timeouts, Retries, and Unknown States](part-3-reliability/timeout-and-retry.md)
* [Chapter 11: Idempotency and Payment Attempts](part-3-reliability/idempotency.md)
* [Chapter 12: Fallback Routing](part-3-reliability/fallback-routing.md)
* [Chapter 13: Circuit Breakers and Provider Health](part-3-reliability/provider-health.md)

## Part IV: Operating at Scale

* [Chapter 14: Routing Decision Storage](part-4-operating-at-scale/routing-decision-store.md)
* [Chapter 15: Routing Observability](part-4-operating-at-scale/observability.md)
* [Chapter 16: Routing Experiments](part-4-operating-at-scale/experimentation.md)
* [Chapter 17: Reconciliation](part-4-operating-at-scale/reconciliation.md)
* [Chapter 18: Settlement and Cost Attribution](part-4-operating-at-scale/settlement.md)

## Part V: Global Routing and Evolution

* [Chapter 19: Multi-Region Payment Routing](part-5-global-routing/multi-region.md)
* [Chapter 20: The Routing Configuration Platform](part-5-global-routing/configuration-platform.md)
* [Chapter 21: Intelligent Routing](part-5-global-routing/intelligent-routing.md)

## Part VI: Integrated Case Study

* [Chapter 22: Designing a Global Payment Routing Platform from Scratch](part-6-case-study/global-payment-routing-platform.md)

## Appendices

* [Appendix A: State Models for Payment Intents, Attempts, and Provider Requests](appendices/payment-state-models.md)
* [Appendix B: Security, Compliance, and Data Boundaries](appendices/security-compliance.md)
* [Appendix C: Capacity Planning and Launch Checklist](appendices/capacity-planning-and-launch-checklist.md)
* [Glossary](glossary.md)
