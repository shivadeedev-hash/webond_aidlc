---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secret Protection Best Practices

**Pack:** ASA Base Pack  
**Description:** Ensure that secrets like credentials remain confidential.

## Domain / Applicability

Applies to systems that use secrets, credentials, API keys, tokens, or other sensitive authentication materials.

## Evaluation Criteria

A compliant system should minimize its use of secrets by leveraging existing authentication mechanisms (like IAM roles or service accounts), store necessary secrets in dedicated management services, have defined rotation policies, and maintain procedures for credential revocation.
To help evaluate compliance, a document should explain what secrets the system uses, how they are stored and rotated, and justify why any new secrets are needed instead of existing authentication mechanisms.
A system is clearly non-compliant if it creates unnecessary new secrets where existing auth mechanisms would suffice, stores secrets insecurely (e.g., in code or config files), lacks rotation policies, or has no process for handling compromised credentials.
