---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Data At Rest Access Control Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define access controls for data at rest that enforce least privilege, prevent unauthorized access, and protect data integrity. The design must specify how access to data resources is controlled through appropriate policies and permissions. The architecture must demonstrate protection mechanisms appropriate for the data sensitivity including versioning for critical data recovery and public access prevention for non-public data. The design must show that access controls are restrictive rather than permissive, granting only necessary permissions to authorized principals.

## Domain / Applicability

Data access control architecture documents, resource policy specifications, and data protection designs. Applies to workloads storing data at rest requiring access controls including S3 objects, databases, file systems, and any data requiring protection from unauthorized access or accidental deletion.

## Evaluation Criteria

Review architecture documents to verify data at rest access control design:
- COMPLIANT: Design specifies access controls for data at rest. Must include: (1) Access control mechanisms (resource policies like S3 bucket policies, IAM policies, KMS key policies, database access controls, or other authorization mechanisms); (2) Least privilege approach (restrictive permissions granting only necessary access, not overly permissive); (3) Protection mechanisms appropriate for data (versioning for critical data recovery if applicable, immutability mechanisms if required, public access prevention for non-public data). Design demonstrates access controls prevent unauthorized access and protect data integrity.
- NON-COMPLIANT: Design allows unrestricted or overly permissive access to data, lacks access control mechanisms, does not prevent public access to non-public data, or has no protection mechanisms for critical data. Missing data at rest access control architecture.
- INSUFFICIENT-DATA: Architecture mentions data storage but does not specify access control mechanisms, lacks details on permissions and policies, does not document protection mechanisms, or fails to show how unauthorized access is prevented.
