---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Third Party Service Provider Security Accountability

**Pack:** PCI DSS Pack  
**Description:** Workloads that process, store, or transmit account data should clearly identify and document all third-party service providers (TPSPs) involved in the data flow, including cloud services, payment processors, tokenization providers, and SaaS integrations. The architecture design should define a shared responsibility model for each TPSP, specifying which security requirements are managed by the TPSP, which by the workload owner, and which are shared, and should define the security controls that apply at the integration boundaries with each TPSP. This covers both pre-engagement selection and due diligence and ongoing operational accountability and monitoring, ensuring accountability boundaries are well-understood and that no security requirements fall through gaps between parties.

## Domain / Applicability

Architecture design phase and ongoing operations for any workload that shares account data with or relies on third-party services that could affect the security of the cardholder data environment (CDE).

## Evaluation Criteria

COMPLIANT: The design documentation identifies all third-party services that process, store, or transmit account data or could affect CDE security. For each TPSP, the design describes the services provided, documents the shared responsibility boundaries (which security controls are managed by the TPSP vs. the workload owner vs. shared), and references written agreements or compliance attestations (such as PCI DSS AOC, SOC 2 reports, or equivalent). The design demonstrates that due diligence was performed in selecting TPSPs, defines the security controls applied at integration boundaries (such as data flow protections, access controls, and monitoring), and includes a plan for periodic compliance monitoring. Alternatively, the design documents that no TPSPs are used for in-scope data handling. NON-COMPLIANT: The design integrates third-party services that handle account data without identifying them, lacks documentation of shared responsibility boundaries, does not reference TPSP compliance status or security agreements, fails to define controls at integration boundaries, or does not address how security requirements are divided between the workload and its TPSPs without documented justification.
