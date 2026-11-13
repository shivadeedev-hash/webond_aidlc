---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Account Data Classification And Retention Controls

**Pack:** PCI DSS Pack  
**Description:** Workloads that store, process, or transmit payment account data should clearly identify and classify cardholder data (CHD) elements—including PAN, cardholder name, expiration date, and service code—and sensitive authentication data (SAD) elements—including full track data, card verification codes, and PINs/PIN blocks. The architecture should enforce that SAD is never retained after transaction authorization, even in encrypted form. PAN, if stored post-authorization, should be rendered unreadable through cryptographic means or equivalent protection. Third-party services handling account data on behalf of the workload should be subject to documented accountability for data protection.

## Domain / Applicability

Any workload that stores, processes, transmits, or could impact the security of payment card account data, including systems that interact with third-party payment processors.

## Evaluation Criteria

COMPLIANT: Design explicitly identifies which account data elements (CHD and SAD) are handled by the workload; architecture demonstrates that SAD is purged immediately after authorization with no persistent storage; PAN storage (if any) includes rendering it unreadable (e.g., tokenization, truncation, hashing, or strong encryption); data flow diagrams show where account data resides and how it transitions; third-party payment service providers are identified with documented responsibility for data protection. NON-COMPLIANT: Design lacks identification of account data elements flowing through the system; architecture permits or does not explicitly prevent SAD storage after authorization; PAN is stored in readable form without documented justification and compensating controls; no data flow mapping exists for payment data; third-party payment processors are used without documented accountability for data protection.
