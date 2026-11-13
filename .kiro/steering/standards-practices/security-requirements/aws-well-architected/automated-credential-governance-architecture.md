---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Credential Governance Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how credentials used by the workload are managed, rotated, and monitored throughout their lifecycle. The design must specify automated mechanisms for rotating long-term credentials (database passwords, API keys, service account credentials, access keys) to limit credential lifespan and reduce exposure risk. The architecture must demonstrate monitoring for credential misuse or anomalous usage patterns. For credentials that cannot be eliminated in favor of temporary credentials (IAM roles), the design must specify rotation schedules, automation approaches, and secure storage mechanisms. The architecture must show how unused or obsolete credentials are identified and removed to maintain least privilege.

## Domain / Applicability

Credential management architecture documents, secrets management designs, application security architecture, and workload authentication specifications. Applies to workloads using long-term credentials including database connections, third-party API integrations, service-to-service authentication, legacy application credentials, and any system requiring management of secrets or API keys.

## Evaluation Criteria

Review architecture documents to verify workload credential management design:
- COMPLIANT: Design specifies automated credential lifecycle management for workload credentials. Must include: (1) Credential inventory identifying long-term credentials used by the workload (database passwords, API keys, service account credentials, access keys); (2) Automated rotation mechanism (AWS Secrets Manager with automatic rotation, custom rotation Lambda functions, or other automated rotation solution) with defined rotation schedules; (3) Secure credential storage (Secrets Manager, Parameter Store with encryption, or other secrets management solution); (4) Monitoring for credential misuse (CloudTrail logging of secret access, CloudWatch alarms for unusual patterns, or anomaly detection); (5) Process for identifying and removing unused credentials. Design shows preference for temporary credentials (IAM roles) over long-term credentials where possible.
- NON-COMPLIANT: Design relies on manual credential rotation, stores credentials in code or configuration files without secure storage, lacks monitoring for credential usage, has no defined rotation schedule for long-term credentials, or does not identify and remove unused credentials. Missing automated credential management.
- INSUFFICIENT-DATA: Architecture mentions credentials, secrets, or authentication but does not specify how credentials are rotated, lacks details on secure storage mechanisms, does not document monitoring for credential misuse, or fails to show how credential lifecycle is managed.
