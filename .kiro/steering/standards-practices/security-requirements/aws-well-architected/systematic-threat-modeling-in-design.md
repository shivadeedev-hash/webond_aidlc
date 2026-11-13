---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Systematic Threat Modeling In Design

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must demonstrate that security threats were considered during design. The design should identify potential threats to the workload and show how architectural decisions and security controls address those threats.

## Domain / Applicability

System architecture documents and security architecture specifications. Applies to workload designs, particularly those handling sensitive data or critical business functions.

## Evaluation Criteria

Review architecture documents to verify threat consideration:
- COMPLIANT: Design demonstrates threat consideration. Shows: (1) Threats identified (potential security threats to the workload); (2) Architectural mitigations (security controls or design decisions addressing threats); (3) Threat-informed design (evidence that threat consideration influenced architecture). Design reflects security threat awareness.
- NON-COMPLIANT: Design shows no consideration of security threats or lacks security controls addressing potential threats. Missing threat consideration.
- INSUFFICIENT-DATA: Architecture mentions security but does not identify specific threats or show how threats influenced design decisions.
