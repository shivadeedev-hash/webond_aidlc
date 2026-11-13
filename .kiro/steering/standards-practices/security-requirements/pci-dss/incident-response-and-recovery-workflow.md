---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Incident Response And Recovery Workflow

**Pack:** PCI DSS Pack  
**Description:** The workload architecture should define how suspected or confirmed security incidents affecting cardholder data environments are responded to and recovered from. The design is expected to specify how alerts from detection mechanisms are routed and escalated, how containment and mitigation actions are supported architecturally, and how data backup and recovery processes support business continuity during incident response. The underlying detection mechanisms (intrusion detection/prevention, file integrity monitoring, payment page tamper detection) are owned by the intrusion-detection-file-integrity-and-page-tamper control; this control consumes their alerts and focuses on the response and recovery workflow rather than the detection capabilities themselves.

## Domain / Applicability

Workloads that process, store, or transmit cardholder data or that operate within or connect to a cardholder data environment (CDE). Applies during architecture design, implementation, and ongoing operations.

## Evaluation Criteria

COMPLIANT: The architecture defines how detection alerts (from the mechanisms owned by the intrusion-detection-file-integrity-and-page-tamper control — IDS/IPS, file integrity monitoring, payment page tamper detection) are routed and escalated to incident response channels. It includes defined containment and mitigation paths, data backup and recovery mechanisms that support business continuity and defined recovery objectives, and a mechanism for detecting unexpected PAN storage. Alternative response or recovery approaches are acceptable when justified and documented with compensating controls.

NON-COMPLIANT: The design has no defined alert routing or escalation path to incident response, lacks a backup/recovery strategy for incident scenarios, or does not address containment and mitigation of confirmed incidents without documented justification. No consideration is given to detecting unexpected PAN storage.
