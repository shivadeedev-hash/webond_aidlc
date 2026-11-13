---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Programmatic Access Control For Cardholder Data Stores

**Pack:** PCI DSS Pack  
**Description:** Access to repositories storing cardholder data (CHD) should be mediated through programmatic methods such as application layers, APIs, or database stored procedures rather than allowing direct query access by end users. Only designated, responsible administrators should have direct access to CHD repositories for administrative duties. Access and allowed actions should be based on user roles and least privilege principles, with separation of duties considered to prevent any single individual from having unchecked end-to-end control over sensitive data processes.

## Domain / Applicability

Workloads that store, process, or transmit cardholder data in databases, file systems, cloud storage, or other data repositories. Applies to the data access architecture and access control design for any persistent CHD storage.

## Evaluation Criteria

COMPLIANT: Design specifies that end-user access to CHD repositories is mediated through application logic, APIs, stored procedures, or equivalent programmatic interfaces with role-based access controls and least privilege enforcement. Design restricts direct database/repository query access to designated administrators only. Design demonstrates separation of duties where individuals performing multiple functions do not have unchecked end-to-end control. NON-COMPLIANT: Design allows end users or application components to perform direct, unfiltered queries against CHD repositories without an intermediary access control layer. Design lacks role-based restrictions on data access actions. Design does not identify or restrict which administrators have direct query access to CHD stores. No separation of duties is considered for sensitive data access paths.
