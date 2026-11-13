---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Certificate Management Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define a certificate management strategy for securing data in transit using Transport Layer Security (TLS) with automated provisioning, deployment, and renewal to prevent certificate expiration incidents. The design must specify how certificates are obtained, deployed to services, and automatically renewed before expiration. The architecture must demonstrate integration with certificate management services or automation mechanisms that eliminate manual certificate lifecycle management. The design must show how automated renewal prevents service disruptions from expired certificates and reduces operational overhead of manual certificate tracking and replacement.

## Domain / Applicability

Certificate management architecture documents, TLS implementation designs, and encrypted communication specifications. Applies to workloads with encrypted network communications including public-facing websites and APIs, internal services requiring TLS, load balancers, API gateways, and any system using certificates for secure communications.

## Evaluation Criteria

Review architecture documents to verify certificate management design:
- COMPLIANT: Design specifies automated certificate management. Must include: (1) Certificate provisioning approach (AWS Certificate Manager, Let's Encrypt with automation, Private CA, or other automated certificate authority); (2) Automated deployment to services (load balancers, API gateways, web servers, application endpoints); (3) Automated renewal mechanism that renews certificates before expiration without manual intervention; (4) Integration with services that use certificates (explicitly shows how certificates are deployed and updated). Design demonstrates elimination of manual certificate lifecycle management and prevention of expiration-related outages.
- NON-COMPLIANT: Design relies on manual certificate management requiring human intervention for renewal, lacks automated renewal mechanisms risking certificate expiration, does not specify how certificates are provisioned or deployed, or requires manual tracking of certificate expiration dates. Missing automated certificate management.
- INSUFFICIENT-DATA: Architecture mentions TLS, HTTPS, or certificates but does not specify how certificates are obtained, lacks details on automated renewal mechanisms, does not document certificate deployment to services, or fails to show how certificate lifecycle is managed.
