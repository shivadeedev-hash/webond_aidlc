---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Strong Cryptography For Cardholder Data In Transit

**Pack:** PCI DSS Pack  
**Description:** Cardholder data (particularly PAN) should be protected with strong cryptography during transmission over any untrusted or public network, including the Internet, wireless, and cellular networks. The workload is expected to enforce current, industry-accepted cryptographic protocols and cipher suites, reject deprecated or insecure protocol versions (such as SSL, SSH v1.0, and early TLS), and ensure that certificates used for transport encryption are valid, trusted, and not expired or revoked. PAN should not be transmitted via end-user messaging technologies (e.g., email, SMS, chat) unless secured with strong cryptography.

## Domain / Applicability

All workload components that transmit, receive, or relay cardholder data across network boundaries, including API endpoints, service-to-service communication channels, database connections, message queues, and any integration with external systems or end-user messaging platforms.

## Evaluation Criteria

COMPLIANT: Design specifies that all data-in-transit paths carrying cardholder data use strong, current cryptographic protocols (e.g., TLS 1.2+ or equivalent) with appropriate cipher suites; design documents or configurations demonstrate that insecure protocol fallback is disabled; certificate management approach is defined including validation, expiration monitoring, and trust chain verification; an inventory or tracking mechanism for keys and certificates protecting PAN transmission is described; and any use of end-user messaging for PAN is addressed with strong encryption or explicitly prohibited. Alternative cryptographic approaches are acceptable when justified with documented rationale and risk assessment.

NON-COMPLIANT: Design lacks specification of transport encryption for cardholder data flows; design permits or does not explicitly prevent fallback to deprecated protocols (SSL, early TLS); no certificate validation or management strategy is documented; PAN transmission paths over public or untrusted networks are identified without corresponding encryption controls; or PAN is sent via end-user messaging technologies without documented strong cryptographic protection.
