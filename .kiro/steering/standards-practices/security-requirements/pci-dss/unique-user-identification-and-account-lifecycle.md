---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Unique User Identification And Account Lifecycle

**Pack:** PCI DSS Pack  
**Description:** The workload should ensure that all human users, service accounts, and application processes are assigned unique identifiers to maintain individual accountability. Shared, group, or generic credentials should be avoided; where exceptional circumstances require them, strict controls including documented justification, management approval, time-limited use, and individual attribution of actions should be enforced. The design should incorporate account lifecycle management covering provisioned access with appropriate approval, prompt revocation upon termination, disabling or removal of inactive accounts within 90 days, and time-limited enablement of third-party remote access accounts with activity monitoring.

## Domain / Applicability

Applies to all workload components that manage user identities, authenticate users or services, or provide access to system components and sensitive data. This includes application-level identity management, service-to-service authentication, administrative access, and third-party integrations.

## Evaluation Criteria

COMPLIANT: Design specifies unique identity assignment for all users and processes (e.g., via an identity provider such as AWS IAM, Azure AD, Okta, or equivalent). Design addresses account lifecycle with mechanisms for provisioning, modification approval, timely deprovisioning of terminated users, and automated detection/disabling of inactive accounts. Shared or generic accounts are either absent or documented as exceptions with compensating controls for attribution and time-limited use. Third-party remote access accounts are designed to be enabled only when needed and monitored. NON-COMPLIANT: Design relies on shared or generic credentials without documented justification and compensating controls. Design lacks a defined account lifecycle process (no provisioning approval, no termination revocation, no inactive account handling). Design does not address individual accountability or action attribution. Third-party access accounts have no time-limiting or monitoring provisions.
