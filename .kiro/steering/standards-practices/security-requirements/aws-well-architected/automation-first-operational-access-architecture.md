---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automation First Operational Access Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define an operational strategy that prioritizes automation over manual interactive access for deployment, configuration, maintenance, and operational tasks. The design must specify how operational activities are performed through automated mechanisms rather than requiring manual intervention or interactive access to resources. The architecture must demonstrate automation for routine operational tasks and show that manual interactive access is minimized, audited, and justified when necessary. The design must show how automation reduces operational risks including human error, unauthorized access, and credential management overhead.

## Domain / Applicability

Operational architecture documents, automation strategy specifications, deployment and configuration management designs, and operational workflow documentation. Applies to workloads with operational tasks requiring deployment, configuration, maintenance, patching, or troubleshooting activities that could be automated rather than performed manually.

## Evaluation Criteria

Review architecture documents to verify automation-first operational approach:
- COMPLIANT: Design specifies automation-first operational strategy. Must include: (1) Automation approach for operational tasks (infrastructure-as-code for deployments, CI/CD pipelines for releases, configuration management tools, automated patching, or managed service automation); (2) Minimized manual access showing interactive access is limited (automated deployments instead of manual changes, programmatic configuration instead of interactive sessions, managed services reducing operational overhead); (3) Audit and justification for manual access when used (logging of manual activities, approval workflows, emergency access procedures, or break-glass scenarios); (4) Risk reduction through automation (reduced human error, eliminated credential sharing, consistent repeatable processes). Design demonstrates preference for automation over manual intervention.
- NON-COMPLIANT: Design relies primarily on manual interactive access for operational tasks, lacks automation for deployments and configuration, requires routine manual intervention for maintenance, or does not minimize manual access risks. Missing automation-first operational strategy.
- INSUFFICIENT-DATA: Architecture mentions operations or deployments but does not specify automation approach, lacks details on how operational tasks are performed, does not document manual access minimization, or fails to show preference for automation over manual intervention.
