---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Comprehensive Audit Logging For Sensitive Data Access

**Pack:** PCI DSS Pack  
**Description:** The workload should implement comprehensive audit logging across all system components that process, store, or transmit sensitive data (such as cardholder data). Audit logs are expected to capture: individual user access to sensitive data, all actions by privileged/administrative accounts, access to audit logs themselves, invalid logical access attempts, changes to authentication credentials and account privileges, audit log lifecycle events (initialization, stopping, pausing), and creation/deletion of system-level objects. Each log entry should include user identification, event type, date and time, success or failure indication, event origin, and the identity of the affected data, component, resource, or service.

## Domain / Applicability

All workload components that handle sensitive data, authentication, authorization, or administrative functions — applicable during architecture design and throughout code implementation of logging instrumentation.

## Evaluation Criteria

COMPLIANT: Design specifies a logging strategy that addresses the key auditable event categories (user data access, privileged actions, failed access attempts, credential changes, audit log access, system object changes) and defines log record fields including user identity, event type, timestamp, success/failure, origin, and affected resource. Implementation demonstrates logging instrumentation in application code and infrastructure configuration. Alternative logging approaches (centralized logging services, application-level libraries, sidecar patterns) are acceptable when they achieve equivalent coverage. NON-COMPLIANT: Design lacks a defined logging strategy for sensitive data access, omits critical event categories (e.g., privileged actions or failed access attempts) without justification, fails to specify required log record details, or shows no mechanism for capturing audit events across system components.
