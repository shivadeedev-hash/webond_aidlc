---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Threat Intelligence Integration

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must specify how threat intelligence and security updates will be integrated into the workload's security posture through automated mechanisms. The design must identify threat intelligence sources and automated services that provide continuous updates on emerging threats, vulnerabilities, and attack patterns. The architecture must demonstrate how threat intelligence informs the workload's detection, prevention, and response capabilities. The design must show preference for automated threat intelligence integration that stays current with evolving threats without requiring manual threat feed management or constant manual updates.

## Domain / Applicability

Security architecture documents, threat detection and response designs, vulnerability management architecture, and security monitoring specifications. Applies to workloads requiring continuous threat awareness including internet-facing applications, data processing systems, and environments handling sensitive information.

## Evaluation Criteria

Review architecture documents to verify threat intelligence integration:
- COMPLIANT: Design incorporates automated threat intelligence integration. Must include: (1) Threat intelligence sources identified (AWS managed services like GuardDuty/Inspector/WAF, third-party threat feeds, vulnerability databases, security bulletins); (2) Automated update mechanism (services that automatically incorporate new threat intelligence without manual intervention); (3) Integration with security controls showing how threat intelligence informs detection rules, prevention controls, or response workflows; (4) Demonstrates staying current with evolving threats through automation rather than manual processes. Acceptable approaches include AWS managed services with automatic updates, integrated threat intelligence platforms, or automated vulnerability scanning with current threat databases.
- NON-COMPLIANT: Design relies solely on manual threat intelligence processes, lacks any automated threat intelligence integration, does not specify how threat information will be obtained or updated, or requires constant manual updates to stay current with threats. Missing automated threat intelligence integration.
- INSUFFICIENT-DATA: Architecture mentions security monitoring or threat detection but does not specify threat intelligence sources, lacks details on automated update mechanisms, or does not document how threat intelligence informs security controls and response capabilities.
