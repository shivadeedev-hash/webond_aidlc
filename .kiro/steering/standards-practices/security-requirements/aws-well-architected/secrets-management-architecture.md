---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secrets Management Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how secrets and credentials are managed securely. The design must demonstrate that credentials are not hardcoded in application code, configuration files, or container images. The architecture must specify use of secure secrets management mechanisms (AWS Secrets Manager, Systems Manager Parameter Store, or equivalent) for storing and retrieving credentials. The design should show preference for temporary credentials (IAM roles) over long-lived credentials where possible.

## Domain / Applicability

Security architecture documents and credential management specifications. Applies to workloads requiring authentication credentials including database connections, API integrations, and service-to-service communication.

## Evaluation Criteria

Review architecture documents to verify secrets management:
- COMPLIANT: Design specifies secure secrets management. Shows: (1) No hardcoded credentials (prohibition of secrets in code, config files, or images); (2) Secrets management service (AWS Secrets Manager, Parameter Store, or equivalent for storing credentials); (3) Secure retrieval (applications retrieve secrets at runtime, not embedded); (4) Preference for temporary credentials (IAM roles used where possible). Design demonstrates credentials are managed securely.
- NON-COMPLIANT: Design allows hardcoded credentials in code or configuration, lacks secrets management service, or fails to specify secure credential handling. Missing secrets management.
- INSUFFICIENT-DATA: Architecture mentions credentials but does not specify secrets management approach, lacks details on secure storage and retrieval, or does not address hardcoding prohibition.
