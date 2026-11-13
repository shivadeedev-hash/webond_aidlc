---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Multi Tenant Isolation And Per Tenant Security

**Pack:** PCI DSS Pack  
**Description:** Multi-tenant workloads that process, store, or transmit sensitive data (such as cardholder data) should implement robust logical separation between tenant environments. The architecture should ensure that tenants cannot access each other's data or resources, that the provider cannot access tenant environments without explicit authorization, and that each tenant is restricted to only its own allocated resources and data. Per-tenant audit logging should be enabled by default and made accessible exclusively to the owning tenant. The design should also incorporate mechanisms to support prompt forensic investigation and secure incident/vulnerability reporting for each tenant.

## Domain / Applicability

Applicable to any workload designed as a multi-tenant service or shared platform where multiple customers share infrastructure, applications, or databases — including SaaS applications, shared hosting environments, payment platforms, and cloud-delivered services handling sensitive data.

## Evaluation Criteria

COMPLIANT: Design demonstrates a clear tenant isolation strategy using logical or physical separation mechanisms (such as separate database schemas/instances per tenant, namespace isolation, row-level security, dedicated compute per tenant, or network segmentation), AND specifies per-tenant access controls ensuring tenants can only reach their own data and resources, AND includes per-tenant audit logging that is active by default and accessible only to the owning tenant, AND describes mechanisms for incident response and forensic investigation support per tenant, AND addresses secure channels for tenants to report security incidents. Alternative isolation approaches are acceptable when justified with documented risk assessment and compensating controls. NON-COMPLIANT: Design lacks a documented tenant isolation strategy, OR does not address how tenants are prevented from accessing each other's data, OR omits per-tenant logging and audit capabilities, OR fails to describe incident response and forensic investigation mechanisms for tenant environments, without documented justification or compensating controls.
