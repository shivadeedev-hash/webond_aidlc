---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Application And System Account Security Management

**Pack:** PCI DSS Pack  
**Description:** Application and system accounts should be strictly managed to ensure they are used only for their intended purpose and are not susceptible to misuse. These accounts pose elevated security risk due to their privileged access (e.g., programmatic database access) and should be designed to prevent interactive login where possible, avoid hard-coded credentials, and protect authentication factors against compromise through rotation and complexity requirements.

## Domain / Applicability

Applicable to any workload that uses service accounts, application accounts, or system-level identities for automated processes, inter-service communication, database access, or background operations.

## Evaluation Criteria

COMPLIANT: Design specifies that application/system accounts are configured to disallow interactive login by default (or documents exceptional circumstances with justification); credentials are retrieved from a secrets management solution (such as AWS Secrets Manager, HashiCorp Vault, Azure Key Vault, or equivalent) rather than hard-coded; account scope is limited to the minimum necessary privileges and constrained to specific machines/services; and a credential rotation strategy is defined. NON-COMPLIANT: Design shows application or system accounts with interactive login enabled without documented justification; credentials are embedded in source code, scripts, or configuration files; no secrets management approach is specified; accounts lack scope restrictions or rotation policies; or no accountability mechanism exists for exceptional interactive use of system accounts.
