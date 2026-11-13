---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cardholder Data Encryption Scope Isolation

**Pack:** PCI DSS Pack  
**Description:** Workloads that handle cardholder data should architect encryption and key management such that encrypted cardholder data is isolated from decryption keys and key management processes. Encryption alone is insufficient to remove cardholder data from PCI DSS scope; the design should ensure that systems storing encrypted cardholder data do not co-reside with or have access to decryption keys. Where third-party service providers are involved, responsibilities for encryption, key management, and applicable security controls should be clearly defined and documented.

## Domain / Applicability

Applicable to any workload that processes, stores, or transmits cardholder data, including architectures that rely on encryption to protect cardholder data at rest or in transit, and designs involving third-party service providers handling encrypted cardholder data.

## Evaluation Criteria

COMPLIANT: The design demonstrates clear architectural separation between systems that store encrypted cardholder data and systems that manage decryption keys (e.g., using dedicated key management services such as AWS KMS, Azure Key Vault, HashiCorp Vault, or equivalent, with access policies restricting key access to only authorized encryption/decryption components). Critically, the design shows that the key-encrypting key (KEK) is stored separately from the data-encrypting key (DEK) it protects — they are not co-located in the same store, service, or environment — and that keys are kept in the fewest possible locations. The design explicitly identifies all components in PCI DSS scope, including encryption/decryption systems and key management systems. Where third-party providers handle encrypted data, the design documents a responsibility matrix clarifying which party owns which security controls. NON-COMPLIANT: The design stores encrypted cardholder data alongside decryption keys or within the same environment without documented justification and compensating controls, OR stores a key-encrypting key alongside the data-encrypting key it protects (even if cardholder data and keys are nominally held in different services). The design lacks identification of which components are in PCI DSS scope due to encryption/key management activities. Third-party responsibilities for encrypted cardholder data handling are undefined or undocumented.
