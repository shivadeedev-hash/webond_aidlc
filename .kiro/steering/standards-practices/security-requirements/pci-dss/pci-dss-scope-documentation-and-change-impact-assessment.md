---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Pci Dss Scope Documentation And Change Impact Assessment

**Pack:** PCI DSS Pack  
**Description:** Workloads that store, process, or transmit payment account data should maintain clear documentation of their PCI DSS scope boundary, including all in-scope systems, networks, and data flows. When changes are made to the workload's architecture or network topology, a formal impact assessment should be performed to determine whether the change affects PCI DSS scope, identify newly applicable requirements, and update scope documentation accordingly. Where network segmentation is used to reduce PCI DSS scope, the segmentation controls should be validated regularly and after any modifications to ensure continued effectiveness.

## Domain / Applicability

Workloads handling payment account data, particularly those processing large volumes or serving as aggregation points for cardholder data. Applies during initial architecture design, ongoing change management, and periodic scope reviews.

## Evaluation Criteria

COMPLIANT: Design documentation clearly identifies the PCI DSS scope boundary, enumerates in-scope components and data flows, and describes a process for assessing PCI DSS impact when architectural changes occur. If segmentation is used, the design specifies segmentation controls and a validation approach (such as periodic penetration testing or equivalent verification). Alternative scoping approaches are acceptable when documented with rationale. NON-COMPLIANT: Design lacks identification of PCI DSS scope boundaries, does not address how architectural changes are evaluated for compliance impact, or uses segmentation to reduce scope without describing how segmentation effectiveness is validated. No documented justification is provided for omissions.
