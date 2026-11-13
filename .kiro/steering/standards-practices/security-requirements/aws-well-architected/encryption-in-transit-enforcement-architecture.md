---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Encryption In Transit Enforcement Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how sensitive data is encrypted in transit to protect confidentiality when transmitted over networks. The design must specify encryption protocols used for data transmission (TLS/HTTPS for web traffic, encrypted connections for databases, VPN for external connectivity). The architecture must demonstrate that sensitive data is protected during transmission using appropriate encryption protocols. The design must show that unencrypted protocols are not used for sensitive data transmission and that modern secure protocols are employed.

## Domain / Applicability

Encryption in transit architecture documents, network security designs, and data transmission security specifications. Applies to workloads transmitting sensitive data including public-facing applications, API communications, database connections, external integrations, and any system requiring encrypted network communications.

## Evaluation Criteria

Review architecture documents to verify encryption in transit implementation:
- COMPLIANT: Design specifies encryption in transit for sensitive data. Must include: (1) Encryption protocols specified (TLS/HTTPS for web traffic and APIs, encrypted database connections, VPN or encrypted links for external connectivity); (2) Protocol versions appropriate (modern TLS versions like 1.2 or 1.3, not outdated versions); (3) Sensitive data protection (demonstrates sensitive data is encrypted during transmission, not sent over unencrypted channels). Design shows appropriate encryption protocols are used for data transmission.
- NON-COMPLIANT: Design transmits sensitive data over unencrypted protocols, lacks encryption for data in transit, uses outdated insecure protocols, or does not specify encryption for sensitive data transmission. Missing encryption in transit implementation.
- INSUFFICIENT-DATA: Architecture mentions data transmission but does not specify encryption protocols, lacks details on how data is protected in transit, or fails to show that sensitive data transmission is encrypted.
