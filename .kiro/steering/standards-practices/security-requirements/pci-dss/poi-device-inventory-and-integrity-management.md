---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Poi Device Inventory And Integrity Management

**Pack:** PCI DSS Pack  
**Description:** Workloads that integrate with Point of Interaction (POI) devices for payment card capture should incorporate capabilities for device inventory tracking and integrity validation. This includes maintaining records of device identifiers (make, model, serial number, location), supporting periodic integrity verification workflows, and enabling detection or alerting of unauthorized device changes or substitutions to reduce the risk of skimming or device tampering attacks.

## Domain / Applicability

Payment processing systems that manage, communicate with, or depend on physical POI devices (card readers, payment terminals) for capturing cardholder data through direct physical interaction.

## Evaluation Criteria

COMPLIANT: Design includes a POI device inventory management capability that tracks device identifiers and locations, supports integrity verification workflows (such as remote attestation, heartbeat monitoring, or integration with inspection scheduling systems), and provides alerting for anomalous device behavior or configuration changes. Alternative approaches are acceptable when justified. NON-COMPLIANT: Design involves POI device integration but lacks any mechanism for device inventory tracking, integrity monitoring, or anomaly detection, without documented justification or compensating controls. NOT-APPLICABLE: This requirement applies only to architectures with physical, card-present POI devices. Per PCI DSS 9.5.1 applicability notes, COTS devices (such as phones and tablets) and components that only accept manual key entry are out of scope. Cloud-native, e-commerce, or tokenized workloads with no physical card-present POI hardware should mark this NOT-APPLICABLE with documented justification.
