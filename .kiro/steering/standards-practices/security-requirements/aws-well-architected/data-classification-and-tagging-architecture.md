---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Data Classification And Tagging Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define a comprehensive data classification scheme that identifies data sensitivity levels, handling requirements, business processes, storage locations, and data ownership for all data the workload processes. The design must document data classification levels aligned with legal and compliance requirements (e.g., Public, Internal, Confidential, Restricted, PHI, PII, PCI) with corresponding data protection requirements for each level. The architecture must specify use of resource tagging to label data stores and resources with classification metadata (Classification tag key with values like PHI, PII; Sensitivity tag key with values like High, Medium, Low), ensuring tags do not contain sensitive data themselves. The design must include a data catalog strategy mapping data across the organization including location, sensitivity level, and controls protecting the data. The architecture must demonstrate AWS Config integration monitoring tagged resources for compliance with protection requirements (encryption settings, access controls) and alerting on mishandling. The design must show tag policies in AWS Organizations defining standard tag keys and acceptable values, enabling determination of whether appropriate controls are in place, can be audited, and appropriate responses exist for data mishandling.

## Domain / Applicability

Data protection architecture documents, data classification policy documentation, tagging strategy specifications, data catalog designs, and compliance monitoring architecture. Applies to all AWS workloads processing or storing data with varying sensitivity levels including regulated data (PHI, PII, PCI), confidential business information, and any data requiring classification-based protection controls.

## Evaluation Criteria

Review architecture documents to verify data classification scheme design:
- COMPLIANT: Design specifies comprehensive data classification architecture. Includes: data classification levels defined with sensitivity levels (Public, Internal, Confidential, Restricted) and regulatory categories (PHI, PII, PCI) aligned with legal and compliance requirements; handling requirements documented for each classification level (encryption, access controls, retention, disposal); data ownership and business processes identified; resource tagging strategy with classification metadata (Classification tag key with values like PHI/PII, Sensitivity tag key with High/Medium/Low values) ensuring tags don't contain sensitive data; data catalog approach mapping data location, sensitivity, and controls; AWS Config integration monitoring tagged resources for compliance (encryption settings, access controls) with alerts for mishandling; tag policies in AWS Organizations defining standard tag keys and acceptable values; demonstrates ability to determine if appropriate controls are in place, can be audited, and appropriate responses exist for mishandling.
- NON-COMPLIANT: Design lacks data classification scheme, does not define sensitivity levels or handling requirements, has no tagging strategy for data identification, lacks data catalog or compliance monitoring, or cannot determine if appropriate controls are in place. Missing data classification architecture.
- INSUFFICIENT-DATA: Architecture mentions data but does not specify classification levels, lacks tagging strategy for data identification, does not document data catalog approach, or fails to show compliance monitoring and auditing of data protection controls.
