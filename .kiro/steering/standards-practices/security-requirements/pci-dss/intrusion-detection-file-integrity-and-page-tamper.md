---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Intrusion Detection File Integrity And Page Tamper

**Pack:** PCI DSS Pack  
**Description:** The workload should implement layered detection mechanisms to identify network intrusions, unauthorized file modifications, and payment page tampering. This includes intrusion detection/prevention at the CDE perimeter and critical internal points, file integrity monitoring for critical system and configuration files, and tamper-detection for payment page scripts and HTTP headers as received by consumer browsers. These mechanisms reduce the risk of undetected compromise, data exfiltration, and web skimming attacks.

## Domain / Applicability

Workloads that process, store, or transmit cardholder data, particularly those with network-accessible CDE components and consumer-facing payment pages.

## Evaluation Criteria

COMPLIANT: Design specifies intrusion detection/prevention capabilities (such as AWS Network Firewall, third-party IDS/IPS, or cloud-native traffic inspection) monitoring traffic at CDE perimeter and critical internal points with alerting to security personnel. Design includes file integrity monitoring (using tools such as OSSEC, Tripwire, AWS Config rules, or equivalent change-detection mechanisms) for critical system files with at least weekly comparison cycles. For workloads with payment pages, design includes a mechanism to detect unauthorized modifications to page scripts and security-impacting HTTP headers as rendered in the consumer browser (using Content Security Policy, Subresource Integrity, real-time page monitoring services, or equivalent tamper-detection approaches). Signatures and detection baselines are kept current. NON-COMPLIANT: Design lacks intrusion detection capabilities at CDE boundaries, does not address file integrity monitoring for critical files, or—for workloads with payment pages—fails to include tamper-detection for page content and headers without documented justification or compensating controls.
