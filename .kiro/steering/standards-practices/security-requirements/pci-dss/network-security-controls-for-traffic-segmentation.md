---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Network Security Controls For Traffic Segmentation

**Pack:** PCI DSS Pack  
**Description:** The workload should implement network security controls (NSCs) to enforce traffic segmentation between environments of different trust levels. NSCs should examine and filter both ingress and egress traffic based on pre-defined policies, ensuring that sensitive data environments are isolated from untrusted networks. Configuration standards for NSC rulesets should be defined, and changes to network configurations should follow a controlled process with periodic review for continued relevance.

## Domain / Applicability

Network architecture and traffic flow design for any workload that processes, stores, or transmits sensitive data, particularly where segmentation between trust zones is required.

## Evaluation Criteria

COMPLIANT: Design specifies network segmentation with explicit trust boundaries, defines ingress and egress filtering policies between zones, documents allowed traffic flows with business justification, and includes a mechanism for periodic review of network rules. Implementation uses network security controls (such as security groups, network ACLs, firewalls, or equivalent) that enforce least-privilege network access. NON-COMPLIANT: Design lacks network segmentation between sensitive and non-sensitive environments, permits unrestricted ingress or egress traffic without justification, has no defined trust boundaries, or provides no process for reviewing and maintaining network security configurations.
