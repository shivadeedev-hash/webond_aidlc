---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Compliance Framework Control Mapping

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must identify applicable compliance requirements, regulatory standards, or security frameworks that apply to the workload based on the data it handles, its business purpose, and regulatory context. The design must document how the workload addresses these requirements through its security controls and architecture decisions. The architecture must demonstrate understanding of which controls are provided by AWS managed services versus which require customer implementation. The design must show how compliance requirements influence architecture choices and control implementation.

## Domain / Applicability

Security architecture documents, compliance requirement specifications, regulatory documentation, and control implementation designs. Applies to workloads with specific compliance obligations including regulated data (healthcare, financial, government), industry standards (PCI-DSS for payments, HIPAA for health data), or organizational security frameworks.

## Evaluation Criteria

Review architecture documents to verify compliance requirements are documented:
- COMPLIANT: Design identifies applicable compliance requirements or standards for the workload. Must include: (1) Compliance requirements identified (regulatory requirements like HIPAA/PCI-DSS, industry standards, organizational security policies, or specific security frameworks); (2) How requirements are addressed (security controls implemented, AWS service capabilities used, customer-managed controls specified); (3) Shared responsibility understanding (which controls are AWS-managed versus customer-managed). Design demonstrates that compliance requirements influenced architecture decisions and control selection.
- NON-COMPLIANT: Design does not identify compliance requirements when workload handles regulated data or has compliance obligations, lacks documentation of how requirements are addressed, or does not show understanding of shared responsibility for controls. Missing compliance requirements documentation.
- INSUFFICIENT-DATA: Architecture mentions compliance or regulatory requirements but does not specify which standards apply, lacks details on how requirements are addressed through architecture, or does not document control implementation approach.
