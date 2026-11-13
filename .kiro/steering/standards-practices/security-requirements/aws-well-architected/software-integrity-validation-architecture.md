---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Software Integrity Validation Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how software artifacts are validated before deployment to ensure integrity and authenticity. The design should specify mechanisms for verifying software artifacts such as container images, code packages, or OS images before use.

## Domain / Applicability

Software deployment architecture documents. Applies to workloads deploying software artifacts including container images, Lambda functions, application code, and OS images.

## Evaluation Criteria

Review architecture documents to verify software integrity validation:
- COMPLIANT: Design specifies software integrity validation. Shows: (1) Artifacts to validate (container images, code, OS images); (2) Validation mechanisms (signature verification, hash checking, or other verification methods); (3) Validation timing (before deployment). Design demonstrates software artifacts are validated.
- NON-COMPLIANT: Design allows deployment of unverified software artifacts or lacks integrity validation mechanisms. Missing software integrity validation.
- INSUFFICIENT-DATA: Architecture mentions software artifacts but does not specify validation approach or lacks details on integrity verification.
