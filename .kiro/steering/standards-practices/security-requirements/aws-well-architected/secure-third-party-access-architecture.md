---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Third Party Access Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define secure third-party access patterns using IAM cross-account roles with temporary credentials rather than long-term IAM credentials. The design must specify use of external IDs in trust policies to prevent confused deputy problems. The architecture must demonstrate least privilege access for third-party roles.

## Domain / Applicability

IAM architecture documents and third-party integration specifications. Applies to workloads with third-party integrations including monitoring solutions, managed services, backup providers, and external parties requiring AWS resource access.

## Evaluation Criteria

Review architecture documents to verify secure third-party access:
- COMPLIANT: Design specifies secure third-party access. Shows: (1) Cross-account IAM roles for third-party access (not long-term credentials like access keys); (2) External IDs in trust policies (to prevent confused deputy); (3) Least privilege (IAM policies scoped to required permissions only). Design demonstrates temporary credentials for third-party access.
- NON-COMPLIANT: Design uses long-term IAM credentials for third-party access, lacks external IDs in trust policies, or grants overly broad permissions. Missing secure third-party access.
- INSUFFICIENT-DATA: Architecture mentions third-party access but does not specify cross-account role approach, lacks external ID usage, or does not show least privilege scoping.
