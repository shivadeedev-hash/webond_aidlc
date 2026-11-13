---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Sad Non Retention After Authorization

**Pack:** PCI DSS Pack  
**Description:** Sensitive Authentication Data (SAD) — including full track data, card verification codes (CVV/CVC), PINs, and PIN blocks — should not be retained in any form after the payment authorization process completes. The workload architecture should ensure SAD is rendered unrecoverable once authorization is finalized. During the brief pre-authorization window where SAD may exist in the system, it should be protected with strong cryptography, preferably using a different key than that used for PAN encryption. This control reduces the risk of counterfeit card generation and fraudulent transactions.

## Domain / Applicability

Payment processing workloads that handle, transmit, or temporarily store sensitive authentication data during card-present or card-not-present transaction flows.

## Evaluation Criteria

COMPLIANT: Design explicitly identifies where SAD enters the system and demonstrates that no persistent storage (databases, file systems, logs, caches, message queues, or backups) retains SAD after authorization completion. Design shows SAD is encrypted with strong cryptography during pre-authorization processing and specifies a mechanism for rendering SAD unrecoverable (e.g., secure deletion, ephemeral in-memory processing, automatic purging). Design uses separate cryptographic keys for SAD and PAN where feasible, or documents rationale for alternative key management. NON-COMPLIANT: Design stores SAD (track data, CVV/CVC, PIN, or PIN block) in any persistent medium after authorization without documented issuer-specific business justification. Design lacks encryption for SAD during pre-authorization. Design does not address SAD data lifecycle or fails to identify SAD touchpoints in the transaction flow. Design logs or caches SAD without purge mechanisms.
