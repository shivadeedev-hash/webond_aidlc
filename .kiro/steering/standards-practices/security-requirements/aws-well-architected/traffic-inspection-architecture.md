---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Traffic Inspection Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define traffic inspection mechanisms where needed to detect threats in network traffic. The design should specify how traffic will be inspected for malicious patterns or policy violations using appropriate security controls.

## Domain / Applicability

Network security architecture documents. Applies to workloads requiring threat detection in network traffic, particularly internet-facing applications.

## Evaluation Criteria

Review architecture documents to verify traffic inspection:
- COMPLIANT: Design specifies traffic inspection where appropriate. Shows: (1) Inspection mechanisms (WAF, Network Firewall, or equivalent for threat detection); (2) Inspection points (where traffic is inspected); (3) Threat detection (how inspection identifies malicious patterns). Design demonstrates appropriate traffic inspection for workload security.
- NON-COMPLIANT: Design lacks traffic inspection for internet-facing workload or fails to specify threat detection mechanisms. Missing traffic inspection.
- INSUFFICIENT-DATA: Architecture mentions security but does not specify traffic inspection approach or lacks details on threat detection mechanisms.
