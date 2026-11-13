---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Pan Display Masking And Storage Protection

**Pack:** PCI DSS Pack  
**Description:** Primary Account Numbers (PAN) should be protected both when displayed and when stored. For display, PAN should be masked so that at most the BIN (first 6-8 digits) and last four digits are visible, with full PAN displayed only to personnel with a documented, legitimate business need. For storage, PAN should be rendered unreadable using field-level or column-level protection mechanisms such as strong cryptography with proper key management, tokenization, truncation, or keyed cryptographic hashing. Disk-level or partition-level encryption alone is not considered sufficient for non-removable media, as it provides transparent decryption upon authentication and does not protect PAN from authorized system users.

## Domain / Applicability

Any workload that processes, displays, transmits, or stores Primary Account Numbers (PAN), including payment processing applications, customer-facing portals, reporting systems, APIs, databases, and any component in the cardholder data environment.

## Evaluation Criteria

COMPLIANT: Design specifies PAN masking for all display contexts (UI, logs, receipts, reports) limiting visibility to BIN and last four digits by default, with role-based access controls governing full PAN display for authorized personnel only. Design demonstrates that stored PAN is rendered unreadable using field-level, column-level, or application-level protection (such as tokenization, strong encryption with key management, truncation, or keyed hashing) rather than relying solely on disk-level encryption for non-removable media. Design addresses prevention of PAN copy/exfiltration through remote-access sessions via technical controls (e.g., clipboard restrictions, DLP mechanisms, or equivalent safeguards). Alternative approaches are acceptable when justified with documented rationale and compensating controls.

NON-COMPLIANT: Design stores PAN in cleartext or relies solely on disk-level or partition-level encryption on non-removable media without additional field-level protection. Design displays full unmasked PAN to all users without role-based restrictions or documented business justification. Design lacks any mechanism to prevent PAN copy/relocation during remote access sessions. Design does not address PAN masking in logs, error messages, or reporting outputs.
