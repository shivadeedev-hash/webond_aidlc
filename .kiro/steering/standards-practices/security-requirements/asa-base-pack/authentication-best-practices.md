---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Authentication Best Practices

**Pack:** ASA Base Pack  
**Description:** Ensure only legitimate users can access the system.

## Domain / Applicability

Applies to systems that distinguish between multiple users or clients, including both human users and programmatic callers.

## Evaluation Criteria

A compliant system should have clearly defined authentication mechanisms that are appropriate for its users/clients.
To help evaluate compliance, a document should describe who/what needs to be authenticated, what authentication methods are used, and how credentials are managed and protected.
A system is clearly non-compliant if it lacks authentication without sufficient justification and guardrails, or uses inappropriate authentication methods like hardcoded credentials.
