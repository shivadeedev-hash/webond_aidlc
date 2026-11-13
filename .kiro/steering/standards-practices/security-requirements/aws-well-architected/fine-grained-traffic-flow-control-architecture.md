---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Fine Grained Traffic Flow Control Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how network traffic is controlled to restrict communications to only necessary flows following least privilege principles. The design must specify network traffic controls that limit connectivity between components and external networks. The architecture must demonstrate appropriate network segmentation and traffic filtering mechanisms. The design must show that traffic is restricted based on requirements, using appropriate controls for the workload's network architecture.

## Domain / Applicability

Network architecture documents, traffic control specifications, and network security designs. Applies to workloads with network connectivity requiring traffic control including VPC-based applications, multi-tier architectures, and any system with network communication between components.

## Evaluation Criteria

Review architecture documents to verify network traffic control design:
- COMPLIANT: Design specifies network traffic controls. Must include: (1) Traffic control mechanisms (security groups, network ACLs, firewall rules, or other network controls); (2) Traffic restrictions (controls limit connectivity to necessary flows, not overly permissive); (3) Network segmentation (appropriate separation of components, tiers, or environments). Design demonstrates traffic is controlled following least privilege principles with appropriate mechanisms for the workload architecture.
- NON-COMPLIANT: Design allows unrestricted network traffic, lacks traffic control mechanisms, uses overly permissive rules without justification, or does not segment network traffic appropriately. Missing network traffic control architecture.
- INSUFFICIENT-DATA: Architecture mentions network connectivity but does not specify traffic control mechanisms, lacks details on traffic restrictions, or fails to show how network traffic is limited to necessary flows.
