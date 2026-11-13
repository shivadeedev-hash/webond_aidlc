---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Authenticated Network Communication Architecture

**Pack:** AWS Well-Architected Pack  
**Description:** Architecture documents must define how network communications are authenticated to verify the identity of communicating parties. The design must specify authentication mechanisms for service-to-service communications (east-west traffic) and external communications (north-south traffic), rejecting the principle that network boundaries alone provide sufficient trust. The architecture must demonstrate authentication through cryptographic protocols (TLS with certificate validation, mutual TLS, IPsec) or identity-based authentication mechanisms (AWS SigV4, IAM role-based authentication). The design must show that services verify caller identity before allowing communications, even when services are within the same network boundary.

## Domain / Applicability

Service communication architecture documents, API security designs, microservices architectures, and any workload with service-to-service or external communications. Applies to workloads with multiple services, API-based systems, microservices, distributed applications, and systems requiring authenticated communications beyond network-based trust.

## Evaluation Criteria

Review architecture documents to verify authenticated communication design:
- COMPLIANT: Design specifies authentication mechanisms for network communications. Must include at least ONE of: (1) Cryptographic authentication: TLS with certificate validation, mutual TLS (mTLS) with X.509 certificates, or IPsec with authentication; (2) Identity-based authentication: AWS SigV4 signing for API calls, IAM role-based authentication for service-to-service calls, or API Gateway with IAM authorization; (3) Zero-trust approach: explicitly documents that services authenticate even within same network boundary, rejecting network-only trust. Design shows how caller identity is verified before allowing communications.
- NON-COMPLIANT: Design relies solely on network boundaries for trust without authentication mechanisms, allows unauthenticated service-to-service communications, or does not verify caller identity before allowing actions. Missing authentication architecture for network communications.
- INSUFFICIENT-DATA: Architecture mentions service communications or APIs but does not specify how caller identity is verified, lacks details on authentication mechanisms (TLS, mTLS, SigV4, IAM), or does not address whether network boundaries alone provide trust.
