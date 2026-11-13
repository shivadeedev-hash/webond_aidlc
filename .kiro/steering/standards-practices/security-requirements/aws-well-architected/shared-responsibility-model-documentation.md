---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Shared Responsibility Model Documentation

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must demonstrate understanding of customer security responsibilities for the AWS services selected. The design should address customer-managed security controls such as OS management (for IaaS), data protection, identity and access management, and network security configurations appropriate to the services used.

## Domain / Applicability

System architecture documents and cloud architecture designs. Applies to designs incorporating AWS services.

## Evaluation Criteria

Review architecture documents to verify customer security responsibilities are addressed:
- COMPLIANT: Design addresses customer security responsibilities for selected services. Shows consideration of: OS management (if using EC2/IaaS), data protection (encryption), IAM configuration, network security controls. Design demonstrates understanding of customer-managed security controls.
- NON-COMPLIANT: Design incorrectly assumes AWS manages customer responsibilities (e.g., assumes AWS manages guest OS patching on EC2) or fails to address customer security controls.
- INSUFFICIENT-DATA: Architecture lists services but does not address customer security responsibilities or lacks clarity on security control management.
