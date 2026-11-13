---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Configuration And System Hardening

**Pack:** PCI DSS Pack  
**Description:** All system components in the workload should be configured according to established hardening standards to reduce the attack surface. This includes changing or removing vendor default credentials and accounts, disabling unnecessary services, protocols, and functions, isolating or securing components with differing security levels, configuring security parameters to prevent misuse, and encrypting all non-console administrative access with strong cryptography. Configuration baselines should be aligned with industry-accepted hardening guidance (such as CIS Benchmarks, NIST guidelines, vendor recommendations, or Cloud Security Alliance standards) and applied before or immediately after components are connected to a production environment.

## Domain / Applicability

Applicable to all system components within the workload, including compute instances, containers, managed services, databases, network devices, and any infrastructure supporting the application. Applies at both initial deployment and throughout the operational lifecycle as new vulnerabilities are identified.

## Evaluation Criteria

COMPLIANT: Design specifies a hardening baseline aligned with recognized standards (e.g., CIS, NIST, vendor guidance) for all system component types. Design addresses removal or change of default credentials and accounts. Architecture demonstrates separation or isolation of functions with differing security levels (e.g., separate compute resources, container isolation, or documented justification for co-location with compensating controls). Design specifies that only required services, ports, and protocols are enabled. Non-console administrative access is designed to use encrypted channels (such as TLS/SSH, VPN, or equivalent encrypted connectivity). Infrastructure-as-code or configuration management tooling is referenced for consistent baseline enforcement. NON-COMPLIANT: Design does not reference or define hardening standards for system components. Default credentials or accounts are not addressed. Components with differing security levels are co-located without isolation or documented justification. Unnecessary services or protocols are enabled without rationale. Administrative access channels lack encryption or the design does not address administrative access security.
