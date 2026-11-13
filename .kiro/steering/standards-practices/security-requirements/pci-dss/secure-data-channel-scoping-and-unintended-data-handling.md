---
inclusion: fileMatch
fileMatchPattern: "**/*"
---

# Secure Data Channel Scoping And Unintended Data Handling

**Pack:** PCI DSS Pack  
**Description:** Workloads that process, store, or transmit sensitive data (such as cardholder data or authentication credentials) should clearly define and document which communication channels and interfaces are intended for sensitive data flows. The design should ensure that sensitive data is only received and processed through secured, explicitly scoped channels. Where sensitive data may be accidentally received via an unintended or insecure channel, the workload should implement measures to either securely delete the data and prevent future receipt on that channel, or bring the channel into scope with appropriate security controls applied.

## Domain / Applicability

Workloads handling sensitive financial data, cardholder data, or authentication credentials — particularly those with multiple ingress points, APIs, messaging queues, or communication interfaces where data could arrive through unintended paths.

## Evaluation Criteria

COMPLIANT: Design clearly documents which communication channels and interfaces are designated for sensitive data receipt and transmission. Design demonstrates that sensitive data flows are restricted to secured channels with appropriate encryption and access controls. Design addresses the scenario of unintended sensitive data receipt with a defined handling strategy (secure deletion or channel inclusion in security scope). Alternative approaches are acceptable when justified and risk-assessed.

NON-COMPLIANT: Design lacks identification of intended sensitive data channels. Design does not address how unintended receipt of sensitive data on insecure channels would be handled. Design shows sensitive data flowing through channels without documented security scoping or justification. No consideration is given to restricting sensitive data to explicitly secured communication paths.
