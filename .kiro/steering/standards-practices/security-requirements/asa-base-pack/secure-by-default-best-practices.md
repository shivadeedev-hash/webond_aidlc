---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure By Default Best Practices

**Pack:** ASA Base Pack  
**Description:** Ensure the system's default configuration is secure.

## Domain / Applicability

Applies to systems with configurable security settings, resource sharing options, or access controls.

## Evaluation Criteria

A compliant system should start with restrictive security settings, require explicit opt-in for less secure configurations, and enable strong security features (like encryption) by default. Resources should be private by default.
To help evaluate compliance, a document should describe the default security configurations, what security features are enabled automatically, and what explicit steps users must take to reduce security levels.
A system is clearly non-compliant if it starts with public sharing enabled, requires users to opt-in to basic security features, or defaults to permissive access controls.
