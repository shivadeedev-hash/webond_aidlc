---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Code Review Process Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define a review and validation process that verifies quality and security of designs and implementations through peer review and analysis. The design must specify when reviews occur (before deployment, before significant changes), who participates (team members, subject matter experts, security reviewers), and the review methodology. The architecture must demonstrate how reviews catch errors, inconsistencies, and security issues that may be overlooked by original authors. The design must show combination of human review providing judgment and context with automated analysis providing consistent validation where applicable, establishing quality gates before changes reach production.

## Domain / Applicability

Development and deployment workflow architecture, review process documentation, quality assurance procedures, and change validation specifications. Applies to workloads with designs or implementations requiring validation including application code, infrastructure-as-code, configuration changes, architecture designs, and any system where peer review helps ensure quality and security before deployment.

## Evaluation Criteria

Review architecture documents to verify review and validation process design:
- COMPLIANT: Design specifies review and validation process. Must include: (1) Review timing (before deployment, before merging changes, before production release); (2) Participants (peer reviewers, subject matter experts, security reviewers - not just original author); (3) Review scope (what is reviewed: code, infrastructure definitions, configurations, architecture designs); (4) Quality gates (reviews required before deployment, approval requirements). May include automated analysis tools (code scanning, infrastructure validation, security checks) complementing manual reviews. Design demonstrates catching errors and security issues through review process.
- NON-COMPLIANT: Design allows changes to reach production without review, has no peer review process, lacks quality gates enforcing reviews, or does not specify who reviews or when reviews occur. Missing review and validation architecture.
- INSUFFICIENT-DATA: Architecture mentions development or deployment but does not specify review process, lacks details on review timing or participants, does not document quality gates, or fails to show how reviews ensure quality before production.
