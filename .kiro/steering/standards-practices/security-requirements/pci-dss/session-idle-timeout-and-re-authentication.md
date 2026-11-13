---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Session Idle Timeout And Re Authentication

**Pack:** PCI DSS Pack  
**Description:** User sessions that access system components or sensitive data (such as cardholder data) should enforce an idle timeout policy that requires re-authentication after a period of inactivity, recommended at 15 minutes or less. This reduces the risk of unauthorized access when a user leaves a session unattended. The timeout and re-authentication mechanism can be applied at the application level, the system level, or both, and may include staged controls (e.g., session lock after 15 minutes, automatic logoff after a longer period) to progressively restrict access over time.

## Domain / Applicability

Applicable to any workload with interactive user sessions that provide access to sensitive data, administrative consoles, or system components within the cardholder data environment. Exceptions may be appropriate for single-transaction terminals or automated processes where session locking is used instead of timeout-based logoff.

## Evaluation Criteria

COMPLIANT: Design specifies a session idle timeout mechanism (e.g., 15 minutes or a risk-justified alternative threshold) that requires re-authentication before the session can be resumed. Design addresses where enforcement occurs (application-level session management, identity provider session policies, system-level screen lock, or equivalent). Staged timeout controls or session lock mechanisms for long-running processes are documented where applicable. Alternative timeout thresholds are acceptable when supported by a documented risk assessment.

NON-COMPLIANT: Design lacks any session idle timeout or re-authentication mechanism for interactive user sessions accessing sensitive data. Sessions remain active indefinitely without inactivity controls and no documented justification or compensating controls are provided. Design fails to address the risk of unattended session access.
