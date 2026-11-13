---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Least Privilege Permission Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must demonstrate least privilege permissions where identities are granted only the minimum access required. The design must specify how permissions are scoped appropriately rather than using overly broad wildcards. The architecture must show that IAM roles and policies grant specific access needed for workload operations.

## Domain / Applicability

IAM architecture documents and permission specifications. Applies to workloads requiring access control.

## Evaluation Criteria

Review architecture documents to verify least privilege:
- COMPLIANT: Design demonstrates least privilege. Shows: (1) Specific permissions (not Action:*, Resource:* without justification); (2) Appropriate scoping (specific actions/resources for workload needs); (3) Minimal access (roles grant only required permissions). Design demonstrates permissions are minimized to necessary access.
- NON-COMPLIANT: Design uses overly broad permissions without justification or grants excessive access. Missing least privilege.
- INSUFFICIENT-DATA: Architecture mentions permissions but does not specify how least privilege is achieved or lacks permission details.
