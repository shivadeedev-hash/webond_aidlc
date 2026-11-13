---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Sensitivity Based Data Protection Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define data protection controls appropriate to the sensitivity of data handled by the workload. The design must identify sensitive data and specify protection mechanisms including encryption and access controls appropriate to the data sensitivity level.

## Domain / Applicability

Data protection architecture documents. Applies to workloads handling sensitive data requiring protection controls.

## Evaluation Criteria

Review architecture documents to verify data protection:
- COMPLIANT: Design specifies data protection appropriate to sensitivity. Shows: (1) Sensitive data identified (what data requires protection); (2) Protection controls specified (encryption at rest/transit, access controls appropriate to sensitivity); (3) Controls aligned with data sensitivity (more sensitive data has stronger controls). Design demonstrates appropriate data protection.
- NON-COMPLIANT: Design lacks data protection, applies no controls to sensitive data, or fails to differentiate protection based on sensitivity. Missing data protection.
- INSUFFICIENT-DATA: Architecture mentions data but does not identify sensitive data, lacks protection controls specification, or does not show controls appropriate to sensitivity.
