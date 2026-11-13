---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Key Management Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how encryption keys are managed securely for data at rest encryption. The design must specify use of a secure key management service (AWS KMS or equivalent) for storing and managing encryption keys. The architecture must demonstrate appropriate access controls for keys and integration with AWS services requiring encryption.

## Domain / Applicability

Key management and encryption architecture documents. Applies to workloads encrypting data at rest including S3 objects, EBS volumes, databases, and other data requiring encryption.

## Evaluation Criteria

Review architecture documents to verify key management:
- COMPLIANT: Design specifies secure key management. Shows: (1) Key management service (AWS KMS or equivalent for storing encryption keys); (2) Access controls (appropriate restrictions on who can use keys); (3) Encryption integration (how keys are used with AWS services for data encryption). Design demonstrates keys are managed securely.
- NON-COMPLIANT: Design lacks key management service, has no access controls for keys, or fails to specify secure key storage. Missing key management.
- INSUFFICIENT-DATA: Architecture mentions encryption but does not specify key management approach, lacks details on key storage and access controls, or does not show encryption integration.
