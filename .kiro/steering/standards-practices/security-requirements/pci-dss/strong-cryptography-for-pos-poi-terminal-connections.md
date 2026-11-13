---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Strong Cryptography For Pos Poi Terminal Connections

**Pack:** PCI DSS Pack  
**Description:** Workloads that communicate with or provide connectivity to POS POI (Point of Interaction) terminals should use strong cryptographic protocols (TLS 1.2 or higher) for all data transmissions. SSL and early TLS (versions prior to TLS 1.2) should not be introduced in new implementations. Where legacy POS POI terminal connections still rely on SSL/early TLS, a documented risk mitigation and migration plan should be maintained to transition to strong cryptography.

## Domain / Applicability

Applicable to workloads that process card-present payment transactions, provide connectivity services to POS POI terminals, or act as intermediaries for terminal communications. Relevant during architecture design, protocol selection, and when integrating with payment terminal infrastructure.

## Evaluation Criteria

COMPLIANT: Design specifies TLS 1.2 or higher for all new POS POI terminal connections, OR for legacy terminal connections using SSL/early TLS, the design includes a documented Risk Mitigation and Migration Plan covering: data flows and systems involved, risk assessment results, vulnerability monitoring processes, change controls preventing new SSL/early TLS introduction, and a migration timeline. NON-COMPLIANT: Design introduces SSL or early TLS for new POS POI terminal implementations without justification, OR existing SSL/early TLS usage lacks a documented migration plan and compensating risk controls, OR the design fails to address cryptographic protocol strength for terminal communications.
