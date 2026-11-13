---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Stored Account Data Protection And Retention

**Pack:** PCI DSS Pack  
**Description:** Workloads that process or store account data (such as PANs, cardholder names, expiration dates, or service codes) should minimize data storage and protect stored data using appropriate methods. Data protection should include rendering PANs unreadable through encryption, truncation, masking, or hashing. The workload should implement data retention controls that limit storage to the minimum amount and duration required by business, legal, or regulatory needs, and should ensure secure deletion or rendering data unrecoverable when retention periods expire. Sensitive authentication data (SAD) should not be stored after transaction authorization is complete.

## Domain / Applicability

Any workload that processes, stores, or transmits payment card account data, including primary account numbers (PANs), cardholder names, expiration dates, service codes, or sensitive authentication data.

## Evaluation Criteria

COMPLIANT: Design documents the locations where account data is stored and specifies protection methods (e.g., encryption, tokenization, truncation, or hashing) for rendering PANs unreadable. Design defines data retention periods with documented business justification, identifies mechanisms for secure deletion or rendering data unrecoverable, and demonstrates that SAD is not retained post-authorization. Design includes a periodic verification process for retention compliance. Alternative approaches (such as tokenization services or third-party vault solutions) are acceptable when they achieve equivalent data protection outcomes.

NON-COMPLIANT: Design stores PANs in readable form without documented justification or compensating controls. Design lacks defined retention periods for account data, has no mechanism for secure deletion, retains SAD after authorization, or fails to identify all storage locations for account data. Design does not address data minimization principles.
