---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Payment Software Lifecycle And Configuration

**Pack:** PCI DSS Pack  
**Description:** Software components that store, process, or transmit cardholder data or sensitive authentication data should be developed and maintained following secure software lifecycle practices, properly configured for the deployment environment, and kept current with vendor-supported versions. This applies to validated payment software, third-party components, and all bespoke or custom software within the cardholder data environment. Customizations to validated payment software should be tracked and assessed for their security impact, and end-of-life software should be replaced or mitigated with compensating controls.

## Domain / Applicability

Workloads that store, process, or transmit payment card data or sensitive authentication data, including the selection, configuration, development, and maintenance of all software components within or impacting the cardholder data environment.

## Evaluation Criteria

COMPLIANT: Design identifies all software components in scope for cardholder data handling and demonstrates: (1) a preference for vendor-supported, current software versions with a plan for lifecycle management; (2) secure configuration practices for payment software beyond default or validated settings; (3) secure software development lifecycle (SSDLC) practices for any bespoke or custom software components; (4) documented tracking of any customizations to validated payment software with associated security review; and (5) vendor accountability requirements when third-party developers build custom software handling cardholder data. Alternative approaches are acceptable when justified with documented risk assessment and compensating controls.

NON-COMPLIANT: Design includes software components handling cardholder data without addressing software currency/lifecycle management, lacks secure development practices for custom or bespoke payment software, relies on end-of-life software without documented justification and compensating controls, or fails to account for the security impact of customizations to validated payment software.
