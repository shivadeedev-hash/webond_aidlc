---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Continuous Security Control Monitoring And Failure Response

**Pack:** PCI DSS Pack  
**Description:** The workload should implement continuous monitoring of its security controls to verify they are operating effectively and as intended. This includes automated detection of security control failures or degradation, alerting responsible parties when anomalies are detected, and defined processes for responding to control failures. Monitoring should cover the full set of critical security controls, including network security, intrusion detection/prevention, change-detection mechanisms, anti-malware protections, access controls, audit logging mechanisms, segmentation controls (where used to reduce scope), audit-log review mechanisms, and automated security testing tools.

## Domain / Applicability

Applicable to all workloads that implement security controls for protecting sensitive data and environments. Relevant during architecture design, operational deployment, and ongoing operations where security controls must remain effective over time.

## Evaluation Criteria

COMPLIANT: Design includes mechanisms for continuously monitoring the health and effectiveness of security controls (e.g., health checks, dashboards, automated tests), with automated alerting when controls fail or degrade. Monitoring covers the full set of critical controls — including network security, intrusion detection/prevention, change-detection, anti-malware, access controls, audit logging mechanisms, segmentation controls (where used), audit-log review mechanisms, and automated security testing tools — so that a failure in any one is detected rather than silently leaving a gap. Design addresses how control failures are detected, reported, and remediated, including promptly restoring the failed control, documenting the duration and cause of the failure, assessing security impact during the outage, root cause analysis, and preventing recurrence. Change management processes include security risk assessment prior to implementation. Alternative monitoring approaches are acceptable when justified and documented.

NON-COMPLIANT: Design lacks mechanisms to detect security control failures or degradation, or monitors only a subset of critical controls while leaving others (such as audit logging or segmentation) unmonitored. No alerting or notification process exists for when security controls stop functioning. Design does not address how control failures would be identified, escalated, or remediated, or lacks a defined response process that restores the control and assesses impact during the failure window. Changes to the environment are not assessed for security impact before implementation.
