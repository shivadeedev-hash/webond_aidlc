---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Public Facing Web Application Attack Protection

**Pack:** PCI DSS Pack  
**Description:** Public-facing web applications should be protected against known web-based attacks through a layered approach. This includes deploying automated inline detection and prevention capabilities (such as a Web Application Firewall) in front of public-facing application endpoints, and conducting periodic application security assessments. Additionally, for applications handling payment or sensitive data entry in the browser, client-side scripts should be authorized, integrity-verified, and inventoried to prevent unauthorized script injection and data exfiltration.

## Domain / Applicability

Applicable to any workload that exposes web applications or APIs to the public internet, especially those handling sensitive data such as payment information. Relevant at both the architectural layer (traffic inspection placement) and the code layer (client-side script management, Content Security Policy).

## Evaluation Criteria

COMPLIANT: Design includes an automated web-based attack detection and prevention solution (such as AWS WAF, Azure Front Door WAF, Cloudflare WAF, or equivalent application-layer filtering) positioned in front of public-facing web endpoints, configured to block or alert on common attack patterns (e.g., OWASP Top 10), with audit logging enabled. Design also addresses client-side script security for sensitive pages through mechanisms such as Content Security Policy (CSP), Subresource Integrity (SRI), script inventories, or equivalent controls to authorize and verify script integrity. Periodic application security assessments (at least annually and after significant changes) are referenced or planned. NON-COMPLIANT: Design exposes public-facing web applications without any inline attack detection or prevention capability and without documented justification for an alternative approach. No consideration is given to client-side script authorization or integrity on pages handling sensitive data. No reference to periodic security assessments of the web application.
