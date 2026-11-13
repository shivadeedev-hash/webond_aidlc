---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Temporary Credentials Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must demonstrate that all authentication mechanisms for human and machine identities use temporary credentials rather than long-term credentials (IAM user access keys). The design must specify how temporary credentials will be obtained through IAM roles, federation, or AWS service-to-service authentication for all AWS API and CLI access patterns. The architecture must show elimination or minimization of IAM users in favor of federated identities for humans and IAM roles for machine identities (EC2 instance profiles, ECS task roles, Lambda execution roles, service-to-service authentication). For any exceptional cases requiring long-term credentials, the design must provide explicit justification and compensating controls. The architecture must demonstrate how temporary credentials with limited lifetimes reduce the risk of credential exposure, theft, or inadvertent disclosure.

## Domain / Applicability

Identity and access management architecture documents, credential management strategy documentation, IAM role design specifications, service authentication architecture, and access pattern diagrams. Applies to all workloads including applications with AWS API access, automated processes, CI/CD pipelines, microservices architectures, and any system requiring authentication to AWS services.

## Evaluation Criteria

Review architecture documents to verify temporary credentials design:
- COMPLIANT: Design eliminates or minimizes IAM users and long-term credentials in favor of temporary credentials. Shows: federated identities or IAM Identity Center for human access (not IAM users); IAM roles for machine identities (EC2 instance profiles, ECS task roles, Lambda execution roles); service-to-service authentication using IAM roles; temporary credential acquisition flows through role assumption (AssumeRole, AssumeRoleWithSAML, AssumeRoleWithWebIdentity); credential lifetime specifications. Documents that AWS API and CLI access uses temporary credentials. Any exceptions requiring long-term credentials include explicit justification and compensating controls. Demonstrates reduced credential exposure risk through time-limited credentials.
- NON-COMPLIANT: Design relies on IAM users with long-term access keys for human or machine identities, uses long-term credentials for AWS API/CLI access without justification, or lacks IAM role architecture for temporary credential generation. Missing strategy to eliminate long-term credentials.
- INSUFFICIENT-DATA: Architecture mentions IAM but does not specify whether temporary or long-term credentials will be used, lacks IAM role design for machine identities, or does not document credential acquisition mechanisms and lifetimes.
