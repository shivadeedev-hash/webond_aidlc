---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Change Management And Environment Separation

**Pack:** PCI DSS Pack  
**Description:** Workloads should implement secure change management practices that include environment separation between pre-production and production, controlled deployment pipelines with approval gates, and safeguards to prevent sensitive production data (such as PANs) from being used in non-production environments. Test data and test accounts should be removed before deployment to production to reduce attack surface and information disclosure risks.

## Domain / Applicability

Applies to all workload environments, CI/CD pipelines, deployment processes, and environment architecture where changes are promoted from development through pre-production to production.

## Evaluation Criteria

COMPLIANT: Design demonstrates clear separation between pre-production and production environments with access controls enforcing the boundary; deployment pipelines include approval gates and security impact assessment steps; architecture prevents use of live sensitive data (e.g., PANs) in non-production environments or documents equivalent CDE-level protections; processes exist to ensure test data and test accounts are removed prior to production deployment. NON-COMPLIANT: Design lacks environment separation or shows no access control enforcement between environments; no deployment approval or review process is defined; live sensitive data is used in pre-production without equivalent security controls; no mechanism exists to prevent test artifacts from reaching production.
