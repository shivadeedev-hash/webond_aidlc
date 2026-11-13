---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Layered Network Segmentation Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define network segmentation that separates workload components based on data sensitivity and access requirements. The design must distinguish between components requiring internet access and those requiring only internal access, establishing appropriate boundaries that restrict unnecessary network pathways. The architecture must demonstrate how segmentation reduces attack surface and limits lateral movement to critical systems and data.

## Domain / Applicability

Network architecture documents, VPC design specifications, and network segmentation documentation. Applies to workloads with network connectivity requiring separation between internet-facing and internal components.

## Evaluation Criteria

Review architecture documents to verify network segmentation:
- COMPLIANT: Design specifies network segmentation. Must include: (1) Segmentation approach (separation of components based on sensitivity/access requirements); (2) Component placement (internet-facing vs internal-only components identified and separated); (3) Network boundaries (controls restricting traffic between segments). Design demonstrates segmentation reduces attack surface and limits lateral movement.
- NON-COMPLIANT: Design places all components without segmentation, mixes internet-facing and internal components inappropriately, or allows unrestricted pathways to critical systems. Missing network segmentation.
- INSUFFICIENT-DATA: Architecture mentions network but does not specify segmentation approach, lacks component placement strategy, or fails to show how segmentation restricts pathways.
