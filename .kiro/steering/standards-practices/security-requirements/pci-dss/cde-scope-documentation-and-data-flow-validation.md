---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Cde Scope Documentation And Data Flow Validation

**Pack:** PCI DSS Pack  
**Description:** Workloads that store, process, or transmit cardholder data should maintain comprehensive documentation of their Cardholder Data Environment (CDE) scope. This includes an up-to-date inventory of all in-scope system components with descriptions of their function, data flow diagrams covering all payment stages and acceptance channels, identification of all locations where account data is stored/processed/transmitted, segmentation controls isolating the CDE, and all third-party connections with access to the CDE. This documentation should be reviewed and confirmed at least annually (or every six months for service providers) and upon any significant change to the in-scope environment.

## Domain / Applicability

Workloads handling payment card data that fall within PCI DSS scope, including systems in the CDE, connected to the CDE, or that could impact CDE security. Applies during initial architecture design, significant environment changes, and periodic scope validation reviews.

## Evaluation Criteria

COMPLIANT: Design documentation includes (1) a clearly defined CDE boundary identifying all in-scope system components with their function/purpose, (2) data flow diagrams showing how account data moves through payment stages and acceptance channels, (3) identification of all storage, processing, and transmission points for account data, (4) documented segmentation controls and the environments from which the CDE is segmented, (5) an inventory of third-party connections with CDE access, and (6) a stated cadence for periodic scope review (at least annually, or every six months for service provider workloads). Alternative documentation formats or tooling are acceptable provided all scope elements are addressed. NON-COMPLIANT: Design lacks a defined CDE boundary, omits data flow diagrams for cardholder data, fails to inventory in-scope components, does not document segmentation controls, or has no process for periodic scope validation — without documented justification for the omission.
