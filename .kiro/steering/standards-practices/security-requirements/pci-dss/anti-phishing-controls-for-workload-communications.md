---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Anti Phishing Controls For Workload Communications

**Pack:** PCI DSS Pack  
**Description:** The workload should implement technical and automated mechanisms to detect and protect against phishing attacks targeting personnel who access in-scope system components. This includes anti-spoofing controls for workload-owned domains and automated filtering or inspection of inbound communications. A combination of approaches is recommended to reduce reliance on human judgment for identifying malicious communications.

## Domain / Applicability

Applicable to any workload that sends emails or communications from its own domain(s), or where personnel access in-scope system components via channels susceptible to phishing (email, messaging, web portals). Relevant during design of communication flows and domain configuration.

## Evaluation Criteria

COMPLIANT: Design specifies anti-spoofing controls for workload-owned domains (such as DMARC, SPF, DKIM or equivalent email authentication mechanisms), AND describes automated mechanisms to detect or block phishing attempts targeting personnel who access the workload (such as email filtering, link inspection, anti-malware scanning, or equivalent protections). Alternative approaches are acceptable when justified with documented rationale and compensating controls. NON-COMPLIANT: Design lacks any anti-phishing technical controls for workload-owned domains or communication channels without documented justification. Design relies solely on personnel awareness training without automated detection or prevention mechanisms.
