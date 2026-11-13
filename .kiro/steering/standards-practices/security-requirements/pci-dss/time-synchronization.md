---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Time Synchronization

**Pack:** PCI DSS Pack  
**Description:** Workload systems should be configured with consistent, accurate time synchronization across all components. Accurate time synchronization is essential for correlating logs, performing forensic analysis, and establishing event sequences. Internal components should receive time only from designated, reliable sources, and access to time configuration settings should be restricted and changes logged.

## Domain / Applicability

Applicable to all workload environments (development, staging, production) where systems generate logs, enforce security controls, or process sensitive data. Particularly critical for production systems handling cardholder data or operating within a compliance boundary.

## Evaluation Criteria

COMPLIANT: (1) Design specifies that all workload components synchronize time from a reliable, consistent source (such as Amazon Time Sync Service, NTP pools, or equivalent time synchronization infrastructure) using UTC or International Atomic Time, with internal components receiving time only from designated sources. (2) Access to time configuration settings is restricted and changes to time settings are logged and reviewed. Alternative time synchronization approaches are acceptable when justified and documented.

NON-COMPLIANT: Design lacks a defined time synchronization strategy, allows individual components to independently source time from arbitrary external servers, or does not restrict and log changes to time configuration on critical systems.
