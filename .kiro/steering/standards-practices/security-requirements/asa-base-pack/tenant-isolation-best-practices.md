---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Tenant Isolation Best Practices

**Pack:** ASA Base Pack  
**Description:** Ensure appropriate separation between system tenants.

## Domain / Applicability

Applies to systems that allow multiple tenants to use shared infrastructure or application instances.

## Evaluation Criteria

A compliant system should have clear isolation boundaries between tenants, dedicated security controls preventing unintended cross-tenant access.
To help evaluate compliance, a document should describe how tenant data and resources are segregated, what prevents tenants from accessing each other's resources, and how shared infrastructure is protected against abuse.
A system is clearly non-compliant if it allows tenants to access other tenants' data in unintended ways, or lacks isolation controls in shared components that would prevent bugs.
