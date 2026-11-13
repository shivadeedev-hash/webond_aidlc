---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Mfa For Cardholder Data Environment Access

**Pack:** PCI DSS Pack  
**Description:** Multi-factor authentication (MFA) should be enforced for all access paths to the Cardholder Data Environment (CDE), including administrative access, non-console access, remote access, and web-based application access. The MFA implementation should require at least two distinct authentication factor types (something you know, something you have, something you are), be resistant to replay attacks, and should not be bypassable without documented and time-limited management authorization. This reduces the risk of unauthorized access through compromised single-factor credentials.

## Domain / Applicability

All access points to CDE components including cloud-hosted systems, on-premises applications, network devices, servers, endpoints, and web-based interfaces that process, store, or transmit cardholder data. Applies to both direct network access and application-layer access.

## Evaluation Criteria

COMPLIANT: Design specifies MFA enforcement at all CDE access points using at least two distinct factor types; design addresses replay attack resistance; design demonstrates that MFA cannot be bypassed without documented exception processes; remote access scenarios show MFA applied at each trust boundary transition (e.g., remote-to-network and network-to-CDE). NON-COMPLIANT: Design lacks MFA for any CDE access path without documented justification; design permits single-factor authentication for administrative or non-console CDE access; design shows MFA implementations susceptible to replay attacks or allows unrestricted bypass; design fails to enforce MFA at trust boundary transitions for remote users accessing the CDE.
