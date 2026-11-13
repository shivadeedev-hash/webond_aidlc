---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cde Network Segmentation And Scope Minimization

**Pack:** PCI DSS Pack  
**Description:** Workloads that store, process, or transmit cardholder data (CHD) or sensitive authentication data (SAD) should implement network segmentation to isolate the cardholder data environment (CDE) from other system components. The architecture should clearly identify all locations and flows of account data, minimize the number of components with access to CHD/SAD, and ensure that out-of-scope components cannot impact the security of cardholder data even if compromised.

## Domain / Applicability

Workloads handling payment card data, including any system components that store, process, transmit, or have connectivity to cardholder data or sensitive authentication data.

## Evaluation Criteria

COMPLIANT: Design clearly identifies the CDE boundary, documents all cardholder data flows, implements network segmentation between CDE and non-CDE components (using network security controls, subnets, security groups, firewalls, or equivalent isolation mechanisms), and demonstrates that non-CDE components cannot reach CDE resources without explicit authorization. Design also shows efforts to minimize data storage locations and restrict account data to as few components as possible. NON-COMPLIANT: Design lacks identification of CDE boundaries, uses a flat network architecture without segmentation for components handling cardholder data, fails to document cardholder data flows, or allows unrestricted connectivity between CDE and non-CDE components without justification or compensating controls.
