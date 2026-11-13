---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Controlled Cross Account Access Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how cross-account or public access to workload resources is controlled and justified when such access is required. The design must specify which resources require external access (cross-account or public), the access mechanisms used (resource policies, IAM role assumption, resource sharing), and controls to ensure access is intentional and appropriately restricted. The architecture must demonstrate that external access is explicitly designed with proper authorization controls rather than inadvertently granted through misconfiguration. For resources requiring public access, the design must document the business justification and security controls applied.

## Domain / Applicability

Cross-account access architecture documents, resource policy specifications, public resource designs, and external access control documentation. Applies to workloads with resources accessed from outside the account including cross-account resource sharing, public-facing resources (websites, APIs, data sharing), and multi-account architectures requiring controlled external access.

## Evaluation Criteria

Review architecture documents to verify cross-account and public access controls:
- COMPLIANT: Design specifies controlled external access when applicable. Must include: (1) Identification of resources with external access (which resources are accessed cross-account or publicly); (2) Access mechanisms specified (resource-based policies like S3 bucket policies, IAM role assumption with trust policies, AWS Organizations resource sharing, or public access configurations); (3) Access controls documented (trust policy restrictions, condition keys limiting access, principal restrictions, or other authorization controls); (4) Justification for external access (business need for cross-account sharing or public availability). Design demonstrates external access is intentional and controlled, not inadvertent.
- NON-COMPLIANT: Design allows uncontrolled public or cross-account access without justification, lacks authorization controls for external access, does not specify which resources have external access, or shows evidence of inadvertent external exposure. Missing controlled external access architecture.
- INSUFFICIENT-DATA: Architecture mentions cross-account or public access but does not specify which resources, lacks details on access mechanisms and controls, does not document authorization restrictions, or fails to show external access is intentional and justified.
