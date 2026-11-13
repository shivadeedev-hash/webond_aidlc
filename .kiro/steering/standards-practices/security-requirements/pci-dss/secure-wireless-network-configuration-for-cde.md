---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Wireless Network Configuration For Cde

**Pack:** PCI DSS Pack  
**Description:** Wireless network environments that connect to or transmit cardholder data should be designed with hardened security configurations. This includes ensuring all vendor default settings (encryption keys, access point passwords, SNMP community strings, and other security-related defaults) are changed or confirmed secure before deployment. Wireless encryption keys and credentials should be resistant to offline brute force attacks and managed through a defined lifecycle process that includes rotation upon personnel changes, suspected compromise, or on a periodic schedule.

## Domain / Applicability

Applicable when the workload architecture includes wireless network connectivity to the Cardholder Data Environment (CDE) or when wireless networks are used to transmit, process, or store account data. Also relevant when workload components communicate over wireless channels in environments subject to PCI DSS.

## Evaluation Criteria

COMPLIANT: Design documents demonstrate that (a) wireless environments connecting to the CDE are identified and hardened with non-default configurations, (b) wireless encryption uses strong, industry-accepted protocols (such as WPA3, WPA2-Enterprise with 802.1X/EAP, or equivalent), (c) a key and credential management process is defined that addresses rotation on personnel departure, role changes, suspected compromise, and periodic refresh, and (d) the design avoids reliance on fixed pre-shared keys or documents compensating controls if pre-shared keys are used. Alternative approaches (such as eliminating wireless connectivity entirely, using certificate-based authentication, or network segmentation isolating wireless from the CDE) are acceptable when documented with rationale. NON-COMPLIANT: Design lacks identification of wireless connectivity to the CDE, relies on vendor default wireless configurations without documented hardening, uses weak or deprecated wireless encryption protocols, or has no defined process for wireless key and credential lifecycle management. NOT-APPLICABLE: The workload has no wireless networks connected to the CDE or used to transmit, process, or store account data. To claim this, the design should provide evidence such as a network and data flow diagram showing no wireless connectivity to the CDE or carrying account data, along with documented justification.
