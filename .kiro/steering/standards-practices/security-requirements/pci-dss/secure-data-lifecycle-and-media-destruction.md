---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Data Lifecycle And Media Destruction

**Pack:** PCI DSS Pack  
**Description:** Workloads handling cardholder data should implement secure data lifecycle management covering storage, classification, backup protection, and destruction. When cardholder data is no longer needed for business or legal purposes, it should be rendered unrecoverable through cryptographic erasure, secure deletion mechanisms, or destruction of underlying storage media. Standard OS-level delete operations are insufficient as they do not prevent data recovery.

## Domain / Applicability

Applies to all workload components that store, process, or back up cardholder data — including databases, file systems, object storage, backups, snapshots, and any persistent storage volumes containing sensitive payment data.

## Evaluation Criteria

COMPLIANT: Design specifies a data retention and destruction strategy for cardholder data, including mechanisms for secure deletion or cryptographic erasure when data is no longer needed. Design addresses backup encryption and access controls for stored cardholder data. Design includes data classification aligned with sensitivity levels. Alternative approaches (e.g., crypto-shredding via key destruction) are acceptable when documented. NON-COMPLIANT: Design lacks any data retention or destruction strategy for cardholder data. Design relies solely on standard delete operations without secure erasure mechanisms. Design does not address backup security or access controls for stored cardholder data. No consideration given to data lifecycle management for sensitive payment information.
