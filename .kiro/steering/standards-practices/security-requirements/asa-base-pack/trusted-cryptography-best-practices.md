---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Trusted Cryptography Best Practices

**Pack:** ASA Base Pack  
**Description:** Ensure only trustworthy cryptographic implementations are used.

## Domain / Applicability

Applies to systems that implement cryptographic operations, including encryption, decryption, hashing, signing, or key management.

## Evaluation Criteria

A compliant system should use established cryptographic libraries, avoid custom implementations, and employ strong modern algorithms with proper key management.
To help evaluate compliance, a document should specify which cryptographic libraries or services will be used and how cryptographic keys will be managed.
A system is clearly non-compliant if it implements custom cryptographic algorithms, uses deprecated or weak algorithms, or handles cryptographic keys in an unsafe manner.
