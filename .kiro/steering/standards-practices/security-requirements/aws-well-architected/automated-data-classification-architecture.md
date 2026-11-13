---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Data Classification Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how data processed or stored by the workload is classified to determine appropriate protection controls. The design must specify the data classification approach used to identify sensitive data types (PII, PHI, financial data, confidential business data) and distinguish them from non-sensitive data. The architecture must demonstrate how data classification is performed, whether through automated scanning, manual tagging, or hybrid approaches. The design must show how classification results drive protection decisions, ensuring sensitive data receives appropriate controls (encryption, access restrictions, logging, retention policies) based on its sensitivity level.

## Domain / Applicability

Data classification architecture documents, data protection strategy specifications, and sensitive data handling designs. Applies to workloads that process or store potentially sensitive data including customer information, healthcare data, financial data, confidential business information, or any data requiring classification to determine appropriate protection controls.

## Evaluation Criteria

Review architecture documents to verify data classification design:
- COMPLIANT: Design specifies data classification approach. Must include: (1) Data inventory identifying types of data processed or stored by the workload; (2) Classification methodology specifying how sensitive data is identified (automated scanning tools like Macie/Comprehend, manual tagging based on data source/type, metadata-based classification, or hybrid approach); (3) Classification categories or sensitivity levels defined (e.g., public, internal, confidential, restricted; or PII/PHI/PCI designations); (4) Protection mapping showing how classification drives security controls (sensitive data gets encryption, access restrictions, audit logging, retention policies; non-sensitive data may have relaxed controls). Design demonstrates that data classification informs protection decisions rather than applying uniform controls regardless of sensitivity.
- NON-COMPLIANT: Design does not identify what types of data are processed, lacks any classification methodology, treats all data uniformly without distinguishing sensitive from non-sensitive data, or does not show how classification drives protection decisions. Missing data classification approach.
- INSUFFICIENT-DATA: Architecture mentions data or storage but does not specify what types of data are handled, lacks classification methodology, does not define sensitivity levels or categories, or fails to show how classification informs protection controls.
