---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Automated Security Testing Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how security requirements are validated and verified before deployment to production. The design must specify the validation approach used to confirm security controls are correctly implemented and functioning as intended. The architecture must demonstrate testing or verification mechanisms that identify security issues before release, whether through automated testing, security reviews, compliance validation, or other verification methods. The design must show how validation is integrated into the development or deployment process to provide early detection of security issues, reducing the risk of deploying systems with security vulnerabilities or misconfigurations.

## Domain / Applicability

Security validation architecture documents, testing strategy specifications, security verification designs, and deployment validation processes. Applies to workloads with security requirements that need validation including custom applications, infrastructure deployments, configuration changes, and any system where security controls should be verified before production deployment.

## Evaluation Criteria

Review architecture documents to verify security validation design:
- COMPLIANT: Design specifies security validation approach. Must include: (1) Security requirements identification (what security controls or properties need validation); (2) Validation mechanism specified (automated security testing, manual security reviews, compliance scanning, configuration validation, penetration testing, or combination of approaches); (3) Integration into deployment process (validation occurs before production deployment, security gates or approval workflows); (4) Issue detection and remediation (how security issues found during validation are addressed before release). Design demonstrates that security controls are verified before deployment, reducing risk of releasing systems with security vulnerabilities or misconfigurations.
- NON-COMPLIANT: Design has no security validation approach, allows deployment to production without verifying security controls, lacks any testing or verification mechanisms, or does not specify how security requirements are validated before release. Missing security validation architecture.
- INSUFFICIENT-DATA: Architecture mentions security controls or deployment but does not specify how security requirements are validated, lacks details on validation mechanisms (testing, reviews, scanning), does not document integration into deployment process, or fails to show how security issues are detected before production release.
