---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cryptographic Key Management For Stored Account Data

**Pack:** PCI DSS Pack  
**Description:** Cryptographic keys used to protect stored account data should be managed through a comprehensive key lifecycle approach that ensures keys are secured against disclosure and misuse. This includes restricting key access to the minimum necessary, maintaining a key hierarchy where key-encrypting keys (KEKs) are at least as strong as data-encrypting keys (DEKs) and stored separately, storing secret/private keys within secure cryptographic devices or in encrypted form, and addressing all lifecycle phases including generation, distribution, storage, rotation, and destruction.

## Domain / Applicability

Any workload that uses cryptography to protect stored account data or sensitive cardholder information, including databases, file systems, and application-level encryption.

## Evaluation Criteria

COMPLIANT: Design documents a cryptographic key management architecture that addresses: (1) key hierarchy with KEKs stored separately from DEKs, with KEKs at least as strong as the DEKs they protect; (2) keys stored in a secure cryptographic device (such as an HSM, AWS KMS, Azure Key Vault, GCP Cloud KMS, or equivalent key management service) or encrypted with a separate KEK; (3) key access restricted to minimum necessary principals; (4) defined key rotation strategy with cryptoperiods; (5) procedures for key compromise response including retirement and replacement; and (6) for service providers, a documented cryptographic architecture describing algorithms, protocols, key strengths, and expiry dates. NON-COMPLIANT: Design lacks a defined key management strategy, stores encryption keys alongside encrypted data without separation, uses keys without defined rotation or expiry, provides no mechanism for key compromise response, or fails to restrict key access to minimum necessary principals without documented justification for the approach taken.
