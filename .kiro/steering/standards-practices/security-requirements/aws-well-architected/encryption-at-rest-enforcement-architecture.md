---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Encryption At Rest Enforcement Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how data at rest is encrypted to protect confidentiality and prevent unauthorized access. The design must specify which data is encrypted, the encryption mechanisms used, and how encryption keys are managed. The architecture must demonstrate that sensitive data is protected through encryption, providing an additional security layer beyond access controls. The design must show appropriate key management practices including key access controls and rotation where applicable.

## Domain / Applicability

Encryption architecture documents, data protection specifications, and key management designs. Applies to workloads storing sensitive data at rest including databases, object storage, file systems, and any data requiring confidentiality protection through encryption.

## Evaluation Criteria

Review architecture documents to verify encryption at rest implementation:
- COMPLIANT: Design specifies encryption at rest for workload data. Must include: (1) What data is encrypted (databases, object storage, file systems, or other data stores); (2) Encryption mechanisms (service-managed encryption like S3 SSE, EBS encryption, RDS encryption; or customer-managed encryption); (3) Key management approach (AWS-managed keys, customer-managed keys with KMS, or other key management); (4) Key access controls (who can use encryption keys, key policies if using CMKs). Design demonstrates sensitive data is protected through encryption and keys are appropriately managed.
- NON-COMPLIANT: Design stores sensitive data without encryption, lacks encryption for data at rest, does not specify encryption mechanisms, or has no key management approach. Missing encryption at rest implementation.
- INSUFFICIENT-DATA: Architecture mentions data storage but does not specify whether data is encrypted, lacks details on encryption mechanisms, does not document key management approach, or fails to show how sensitive data is protected through encryption.
