---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Security Design Principles Integration

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must demonstrate integration of security design principles into the workload design. The design should show how security is considered throughout the architecture including identity management, monitoring and traceability, layered security controls, data protection, and appropriate access patterns.

## Domain / Applicability

System architecture documents and security architecture specifications. Applies to all workload types including web applications, data platforms, microservices architectures, and cloud deployments.

## Evaluation Criteria

Review architecture documents to verify security design principles integration:
- COMPLIANT: Design demonstrates integration of security principles with specific architectural decisions. Shows consideration of: identity and access management, monitoring/logging for traceability, layered security controls, data protection mechanisms, and appropriate access patterns. Design reflects security-conscious architecture.
- NON-COMPLIANT: Design contradicts security principles (e.g., requires long-term credentials, lacks security controls, no monitoring) or shows no security consideration in architecture decisions.
- INSUFFICIENT-DATA: Architecture mentions security but lacks specific architectural decisions showing how security principles are integrated into the design.
