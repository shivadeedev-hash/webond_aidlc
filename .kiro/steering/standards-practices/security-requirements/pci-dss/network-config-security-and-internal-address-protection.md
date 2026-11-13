---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Network Config Security And Internal Address Protection

**Pack:** PCI DSS Pack  
**Description:** Network security control (NSC) configurations — including infrastructure as code templates, scripts, firewall rules, and related settings — should be secured against unauthorized access and modification, and kept consistent with active network configurations. Additionally, the workload architecture should minimize the disclosure of internal IP addresses and routing information to unauthorized parties. Public-facing components should be explicitly authorized for public access and placed in dedicated network segments (such as a DMZ) separated from sensitive internal systems via network security controls.

## Domain / Applicability

Workload network architecture design, infrastructure as code definitions, network security control configuration management, and any component with public-facing exposure.

## Evaluation Criteria

COMPLIANT: The design demonstrates that (1) network security control configurations (IaC templates, firewall rules, security group definitions, etc.) are stored with access controls limiting modification to authorized personnel, and version-controlled to maintain consistency with active configurations; (2) internal IP addresses and routing details are obscured from external exposure using techniques such as NAT, reverse proxies, private subnets, load balancers, or equivalent mechanisms; (3) publicly accessible components are explicitly identified, justified, and architecturally separated from internal/sensitive systems through network segmentation (e.g., public subnets, DMZ patterns, or equivalent isolation). Alternative approaches are acceptable when documented with compensating controls. NON-COMPLIANT: The design lacks access controls or version management for network configuration artifacts, OR internal IP addresses and network topology are exposed to untrusted networks without justification, OR publicly accessible components are not explicitly identified and lack network segmentation from sensitive internal systems, with no documented rationale for the deviation.
