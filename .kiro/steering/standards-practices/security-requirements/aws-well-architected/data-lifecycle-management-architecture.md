---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Data Lifecycle Management Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define a data lifecycle management strategy that addresses how data is retained, backed up, and eventually destroyed based on the workload's requirements. The design must specify retention periods appropriate for the data types and compliance needs, backup approaches to protect against data loss, and data destruction methods when data reaches end of life. The architecture must demonstrate that lifecycle management is automated where possible to reduce manual intervention and ensure consistent application of policies. The design must show how lifecycle requirements are implemented through appropriate mechanisms for the data services used.

## Domain / Applicability

Data lifecycle management architecture documents, retention policy specifications, backup strategy designs, and data destruction procedures. Applies to workloads storing data with lifecycle requirements including retention periods, backup needs, compliance obligations, or eventual destruction across various data services.

## Evaluation Criteria

Review architecture documents to verify data lifecycle management strategy:
- COMPLIANT: Design specifies data lifecycle management strategy. Must include: (1) Retention requirements (how long data is kept, based on compliance, operational, or business needs); (2) Backup approach (how data is backed up to protect against loss, backup frequency, backup retention); (3) Data destruction (how data is deleted when no longer needed, secure deletion methods if required); (4) Lifecycle automation (automated policies, scheduled processes, or other mechanisms reducing manual intervention). Design demonstrates lifecycle management is implemented through appropriate mechanisms for the data services used (lifecycle policies, backup solutions, automated deletion, or other approaches).
- NON-COMPLIANT: Design has no lifecycle management strategy, lacks retention or backup policies, does not address data destruction, or relies entirely on manual lifecycle management without automation. Missing data lifecycle management architecture.
- INSUFFICIENT-DATA: Architecture mentions data storage but does not specify retention requirements, lacks backup approach, does not document data destruction methods, or fails to show how lifecycle management is implemented.
