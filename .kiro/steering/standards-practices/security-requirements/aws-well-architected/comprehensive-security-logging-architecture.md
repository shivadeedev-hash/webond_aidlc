---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Comprehensive Security Logging Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define a logging strategy that captures security-relevant events and operational activities for the workload to support audit, investigation, and monitoring purposes. The design must specify what events are logged (API activity, access attempts, application events, security events), where logs are stored, and retention periods appropriate for the workload's requirements. The architecture must demonstrate how logs support security investigations and compliance needs by providing visibility into workload activities. The design must show that logs are protected from unauthorized modification or deletion and are available when needed for analysis.

## Domain / Applicability

Security logging architecture documents, observability design specifications, log management strategies, and audit trail documentation. Applies to workloads requiring security monitoring, audit trails, compliance evidence, incident investigation capabilities, or operational visibility into system activities.

## Evaluation Criteria

Review architecture documents to verify logging strategy:
- COMPLIANT: Design specifies logging strategy for the workload. Must include: (1) What is logged (API activity, access logs, application events, security events, or other relevant activities); (2) Where logs are stored (CloudWatch, S3, external logging systems, or other destinations); (3) Retention periods (how long logs are kept, aligned with compliance or operational needs); (4) Log protection (access controls, immutability, or other protections against unauthorized modification/deletion). Design demonstrates logs support security investigations and compliance by providing visibility into workload activities.
- NON-COMPLIANT: Design has no logging strategy, does not specify what events are logged, lacks log storage or retention approach, provides no log protection, or cannot support security investigations due to insufficient logging. Missing logging architecture.
- INSUFFICIENT-DATA: Architecture mentions logging but does not specify what is logged, lacks details on log storage destinations, does not document retention periods, or fails to show how logs support security and compliance needs.
