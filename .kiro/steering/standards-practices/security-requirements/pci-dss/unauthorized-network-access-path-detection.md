---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Unauthorized Network Access Path Detection

**Pack:** PCI DSS Pack  
**Description:** The workload's network architecture should account for the detection and management of unauthorized network access paths, including wireless access points. An inventory of authorized wireless access points (if any) should be maintained with documented business justification. The design should incorporate mechanisms to detect rogue or unauthorized network entry points that could bypass established security controls.

## Domain / Applicability

Network environments where the workload processes, stores, or transmits sensitive data, particularly where physical or hybrid infrastructure components exist that could be susceptible to unauthorized wireless access point installation.

## Evaluation Criteria

COMPLIANT: Design documents an inventory of authorized wireless access points (or explicitly states none are used), includes mechanisms for detecting unauthorized network access paths (such as wireless scanning, NAC, wireless IDS/IPS, or equivalent monitoring), and specifies periodic testing at least quarterly. For cloud-native workloads, design acknowledges that physical wireless security is managed by the cloud provider and documents any additional controls for hybrid components. NON-COMPLIANT: Design fails to address unauthorized network access path detection, lacks an inventory of authorized wireless access points in environments where physical infrastructure exists, or provides no mechanism for periodic detection of rogue access points without documented justification for the omission.
