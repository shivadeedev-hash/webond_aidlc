---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Network Architecture And Data Flow Documentation

**Pack:** PCI DSS Pack  
**Description:** The workload architecture should maintain accurate network diagrams showing all connections to and from sensitive data environments, and data-flow diagrams illustrating how sensitive data (such as cardholder/account data) traverses systems, networks, and storage. All exposed services, protocols, and ports should be explicitly identified with documented business justification, and insecure protocols should be mitigated or replaced. This ensures the security boundary is well-understood and prevents unintended exposure of sensitive data through overlooked connections or unnecessary services.

## Domain / Applicability

Architecture design and documentation for workloads that process, store, or transmit sensitive data, particularly cardholder data environments. Applies during initial design, architecture reviews, and whenever changes are made to network connectivity or data flows.

## Evaluation Criteria

COMPLIANT: Design documentation includes (1) a network diagram identifying all connections to/from the sensitive data environment including management and security service connections, (2) a data-flow diagram showing where sensitive data is received, processed, stored, transmitted, and sent externally, and (3) an explicit inventory of allowed services, protocols, and ports with business justification, where insecure protocols are either avoided or have documented compensating controls. NON-COMPLIANT: Design lacks network or data-flow documentation, fails to identify all connection points for sensitive data, includes open or unjustified services/ports, or uses insecure protocols (such as FTP, Telnet, or unencrypted HTTP for sensitive data) without documented risk mitigation or compensating controls.
