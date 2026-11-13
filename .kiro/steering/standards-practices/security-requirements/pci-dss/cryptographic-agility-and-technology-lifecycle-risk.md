---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cryptographic Agility And Technology Lifecycle Risk

**Pack:** PCI DSS Pack  
**Description:** The workload design should document and maintain an inventory of all cryptographic cipher suites, protocols, and key technologies in use, including their purpose and where they are applied. The architecture should be designed for cryptographic agility — the ability to update or replace cryptographic algorithms, protocols, and related technologies without requiring a full redesign. The design should also account for the lifecycle status of all hardware and software technologies, ensuring they remain actively supported, receive timely security patches, and do not introduce compliance gaps. A documented plan for responding to anticipated cryptographic vulnerabilities and for remediating end-of-life technologies is recommended.

## Domain / Applicability

Applicable to any workload that processes, stores, or transmits sensitive data (such as cardholder data), particularly during initial architecture design, technology selection, and periodic architecture reviews. Relevant whenever cryptographic protocols (e.g., TLS versions, encryption algorithms) or third-party technologies are chosen for the workload.

## Evaluation Criteria

COMPLIANT: The design documentation includes all four of the following, each backed by a concrete artifact a reviewer can point to: (1) Cryptographic inventory — a list of every cipher suite and protocol version in use (for example TLS 1.2, TLS 1.3, AES-256), where each is used, and its purpose; a statement that "an inventory exists" is not sufficient. (2) Agility evidence — at least one named architectural pattern that allows algorithms or protocols to change without a code change, such as centralized key and certificate management through a managed service (for example AWS KMS, AWS ACM, Azure Key Vault, or HashiCorp Vault), a centralized TLS termination point (such as a load balancer or API gateway) where protocol and cipher settings are configuration-driven, or a cryptographic abstraction layer that allows algorithm changes without modifying application code. (3) Lifecycle status — the major runtimes, frameworks, operating systems, and managed services listed with their support or end-of-life dates, not merely a statement that lifecycle is considered. (4) Response plan — a plan that names who owns it, how a newly discovered weak algorithm is identified, and the migration and rollback steps. Alternative documentation approaches are acceptable when they provide an equivalent, reviewer-checkable artifact for each of the four elements. NON-COMPLIANT: The design is missing any of the four elements above, or provides only general assurances (for example "the architecture supports cryptographic agility" or "a plan exists") without a specific artifact a reviewer can confirm in the design document, without documented justification for the omission.
