---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Least Privilege Access Control With Default Deny

**Pack:** PCI DSS Pack  
**Description:** The workload should implement an access control model based on least privilege and need-to-know principles. Access to system components, data resources, and cardholder data should be restricted to the minimum permissions required for each user, application, or system account to perform its designated function. The access control system should enforce a default-deny posture, granting access only when explicitly authorized based on defined roles and job classifications.

## Domain / Applicability

Applicable to all workload components that manage access to sensitive data or system resources, including application-level authorization, IAM policies, service accounts, database access controls, and API authorization layers. Relevant during architecture design, implementation, and ongoing access reviews.

## Evaluation Criteria

COMPLIANT: Design defines an access control model with role-based or attribute-based access control (RBAC/ABAC) that maps permissions to job functions or application needs; IAM policies, service account permissions, and application authorization follow least privilege with default-deny; application and system accounts are scoped to only the specific resources and actions required for their function; design includes a mechanism or process for periodic access review. NON-COMPLIANT: Design lacks a defined access control model or uses overly broad permissions (e.g., wildcard policies, admin-level access for routine functions) without documented justification; no default-deny posture is evident; service or application accounts have access beyond what is necessary for their specific function; no consideration for access review or privilege management is documented.
