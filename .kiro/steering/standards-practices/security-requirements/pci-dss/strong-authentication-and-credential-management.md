---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Strong Authentication And Credential Management

**Pack:** PCI DSS Pack  
**Description:** Workloads should establish and enforce strong authentication for all user and administrator access to system components. This includes requiring authentication via recognized factors (knowledge, possession, or inherence), protecting credentials with strong cryptography during transmission and storage, enforcing account lockout policies after repeated failed attempts, and implementing password complexity and history requirements. Where passwords are the sole authentication factor, the design should either enforce periodic rotation or implement dynamic risk-based access analysis.

## Domain / Applicability

Applicable to all workload components that authenticate users or administrators, including application login flows, API authentication endpoints, administrative consoles, and any service that manages or validates user credentials.

## Evaluation Criteria

COMPLIANT: Design specifies an authentication mechanism using at least one recognized authentication factor (knowledge, possession, or inherence) with strong cryptographic protection of credentials in transit and at rest. Design addresses account lockout after no more than 10 invalid attempts with a minimum 30-minute lockout duration or identity re-verification. Password policies enforce a minimum of 12 characters (or 8 if system-constrained) with alphanumeric complexity, password history of at least 4 prior passwords, unique initial/reset values with forced change on first use, and either 90-day rotation or dynamic risk-based access controls for single-factor implementations. Use of a managed identity provider (such as Amazon Cognito, Auth0, Okta, Azure AD, or equivalent) that enforces these policies is an acceptable approach. Alternative implementations are acceptable when documented with equivalent security posture.

NON-COMPLIANT: Design lacks specification of authentication mechanisms or credential protection. Design permits unlimited authentication attempts without lockout. Password policies do not address minimum complexity, history, or rotation/risk-analysis requirements. Credentials are transmitted or stored without strong cryptographic protections. No documented justification is provided for deviations from recommended authentication controls.
