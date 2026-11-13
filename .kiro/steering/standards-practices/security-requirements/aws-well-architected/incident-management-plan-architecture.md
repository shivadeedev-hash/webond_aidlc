---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Incident Management Plan Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must include a comprehensive incident management plan serving as the foundation for incident response program and strategy, providing structured processes for identifying, remediating, and responding to security incidents in a timely manner. The design must align response and recovery strategies with business outcomes and compliance requirements, incorporating relevant frameworks (NIST SP 800-61 for FedRAMP workloads, data residency requirements for PII). The architecture must demonstrate understanding of AWS Shared Responsibility Model for defense-in-depth incident response, documenting customer responsibilities for security in the cloud. The plan must address cloud-specific incident response considerations following AWS Security Incident Response Guide, including AWS-specific operational roles, cloud-native response capabilities, and integration with AWS services. The design must show the incident management plan as a living document with continuous iteration processes to remain current with evolving cloud operations, workload changes, and threat landscape.

## Domain / Applicability

Incident response planning documentation, incident management strategy specifications, compliance-aligned response procedures, and AWS-specific incident response architecture. Applies to all AWS workloads requiring incident response capabilities including regulated environments (FedRAMP, HIPAA, PCI-DSS), workloads storing PII, mission-critical systems, and any organization needing structured incident response processes.

## Evaluation Criteria

Review architecture documents to verify incident management plan design:
- COMPLIANT: Design includes comprehensive incident management plan as foundation document. Contains: structured processes for identifying, remediating, and responding to security incidents; response and recovery strategies aligned with business outcomes and compliance requirements; compliance framework integration (NIST SP 800-61 for FedRAMP, data residency for PII workloads); AWS Shared Responsibility Model documentation showing customer responsibilities for security in cloud; cloud-specific response considerations from AWS Security Incident Response Guide; AWS-specific operational roles and cloud-native capabilities; continuous iteration process for plan updates as workloads and threats evolve; demonstrates plan enables timely response, mitigation, and recovery from security incidents.
- NON-COMPLIANT: Design lacks incident management plan, has no structured incident response processes, does not align with compliance requirements, fails to address AWS Shared Responsibility Model, or has no plan iteration process. Missing incident management plan foundation.
- INSUFFICIENT-DATA: Architecture mentions incident response but does not include comprehensive incident management plan, lacks compliance alignment, does not document AWS-specific response considerations, or fails to show continuous iteration process.
