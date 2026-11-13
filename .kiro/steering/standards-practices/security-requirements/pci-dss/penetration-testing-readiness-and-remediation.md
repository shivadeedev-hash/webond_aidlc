---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Penetration Testing Readiness And Remediation

**Pack:** PCI DSS Pack  
**Description:** The workload architecture should be designed to support regular internal and external penetration testing at both application and network layers. The design should clearly define the scope boundaries (CDE perimeter and critical systems), identify segmentation controls that require validation, and establish a documented approach for assessing and remediating exploitable vulnerabilities discovered during testing. This ensures the workload can be effectively tested and that security weaknesses are addressed in a risk-prioritized manner.

## Domain / Applicability

Workloads processing, storing, or transmitting sensitive data (e.g., cardholder data) where penetration testing is required to validate security controls, segmentation boundaries, and application/network layer defenses. Applies during initial design, after significant infrastructure or application changes, and on an ongoing periodic basis.

## Evaluation Criteria

COMPLIANT: Design documents clearly identify the workload's security perimeter and critical components in scope for penetration testing; segmentation controls (if used) are documented with testable boundaries; the architecture supports both internal and external testing perspectives; application-layer and network-layer attack surfaces are identified; and a remediation workflow is defined for addressing discovered vulnerabilities with risk-based prioritization. NON-COMPLIANT: Design lacks identification of testable security boundaries; segmentation controls are undocumented or untestable by design; no consideration is given to how penetration testing will be conducted against the workload; or no remediation process is defined for addressing findings from security testing.
