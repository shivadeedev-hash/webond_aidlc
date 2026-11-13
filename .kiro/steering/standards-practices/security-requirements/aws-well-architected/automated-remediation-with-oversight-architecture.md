---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Remediation With Oversight Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how the workload automatically enforces configuration requirements and corrects policy violations with appropriate oversight mechanisms. The design must specify which configuration or policy violations can be automatically corrected versus those requiring human judgment. The architecture must demonstrate automated enforcement mechanisms that detect and remediate drift from desired state, such as restoring required configurations, removing unauthorized changes, or enforcing security policies. The design must include oversight controls to ensure automated corrections are appropriate, including validation before execution, monitoring of correction outcomes, and safeguards against unintended impacts. The architecture must show how automated enforcement reduces response time to violations while maintaining control over potentially disruptive changes.

## Domain / Applicability

Automated enforcement architecture documents, self-healing system designs, configuration management specifications, and policy enforcement mechanisms. Applies to workloads with configuration requirements that can drift, systems requiring consistent policy enforcement, applications with self-healing capabilities, and any workload where automated correction of violations reduces risk or improves reliability.

## Evaluation Criteria

Review architecture documents to verify automated enforcement design:
- COMPLIANT: Design specifies automated enforcement with appropriate oversight. Must include: (1) Violation detection mechanism that identifies configuration drift or policy violations (monitoring, validation checks, comparison to desired state); (2) Classification of violations into auto-correctable (standard fixes with low risk) versus requiring human review (complex situations, potential service impact); (3) Automated correction mechanism specified (Lambda functions, automation scripts, configuration management tools, infrastructure-as-code enforcement) with appropriate permissions; (4) Oversight controls including at least ONE of: validation/testing before applying corrections, approval workflows for high-risk changes, monitoring of correction outcomes with alerting, or rollback capabilities for failed corrections. Design demonstrates reduced response time while maintaining control over changes.
- NON-COMPLIANT: Design relies solely on manual correction of violations, lacks any automated enforcement mechanisms, does not distinguish between auto-correctable and human-review situations, or has no oversight controls for automated corrections risking unintended impacts. Missing automated enforcement architecture.
- INSUFFICIENT-DATA: Architecture mentions configuration management or policy enforcement but does not specify how violations are detected, lacks details on automated correction mechanisms, does not document oversight controls, or fails to show classification of what can be auto-corrected versus requiring human judgment.
