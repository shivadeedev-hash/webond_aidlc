---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cde Network Segmentation And Traffic Control

**Pack:** PCI DSS Pack  
**Description:** The workload architecture should enforce strict network segmentation around the cardholder data environment (CDE), implementing deny-by-default traffic controls for both inbound and outbound flows. Network security controls (such as firewalls, security groups, network ACLs, or equivalent mechanisms) should ensure that only explicitly authorized and necessary traffic is permitted to and from components that process or store cardholder data. Systems storing cardholder data should not be directly accessible from untrusted networks, and anti-spoofing measures should be in place to prevent forged source addresses from reaching trusted network segments.

## Domain / Applicability

Workload network architecture design, particularly for any workload that processes, stores, or transmits cardholder data or operates within a CDE. Applies during initial architecture design, network topology decisions, and ongoing architecture reviews.

## Evaluation Criteria

COMPLIANT: The design demonstrates clear network segmentation isolating CDE components from untrusted networks, with deny-by-default inbound and outbound traffic rules. The architecture shows that cardholder data stores are not directly reachable from untrusted networks (e.g., placed in private subnets or equivalent isolated segments). Traffic flows are explicitly documented showing only necessary communications are permitted. The design includes anti-spoofing protections and controls traffic between trusted and untrusted network boundaries using network security controls (such as cloud-native firewalls, security groups, NACLs, WAFs, or equivalent network filtering mechanisms). Internal IP addresses and routing details are not exposed to unauthorized parties.

NON-COMPLIANT: The design lacks network segmentation around CDE components, or cardholder data stores are directly accessible from untrusted networks (e.g., public internet) without documented justification and compensating controls. The architecture uses allow-by-default traffic rules or fails to restrict outbound traffic from the CDE. No anti-spoofing measures are addressed. Traffic flows between trusted and untrusted zones are not controlled or documented.
