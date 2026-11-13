---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Log Protection Best Practices

**Pack:** ASA Base Pack  
**Description:** Protect the integrity and confidentiality of system logs.

## Domain / Applicability

Applies to systems that handle sensitive information that may need to be redacted from logs.

## Evaluation Criteria

A compliant system should protect logs from unauthorized access or tampering, redact sensitive information from logs, retain logs for appropriate durations, and ensure log storage is reliable and secure.
To help evaluate compliance, a document should describe how logs are protected, what information is redacted, how long logs are retained, and who can access them.
A system is clearly non-compliant if it stores logs insecurely, fails to redact sensitive data, lacks appropriate retention policies, or allows unauthorized access to logs.
