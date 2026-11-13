---
inclusion: always
---

# AWS Well-Architected Framework Context

## Core Priorities (in order of importance)

1. **Security** - Always the highest priority
2. **Reliability** - Critical second priority
3. **Performance** - Important third priority

## Security Guidelines

- Apply the principle of least privilege for all IAM roles and policies
- Implement defense in depth with multiple security controls
- Encrypt data at rest and in transit
- Use AWS security services (GuardDuty, Security Hub, WAF, etc.)
- Implement secure VPC designs with proper network segmentation

## Reliability Guidelines

- Design for failure and self-healing
- Implement high availability across multiple AZs
- Use managed services when possible
- Implement proper health checks, monitoring, and alerting
- Test recovery procedures regularly

## Performance Guidelines

- Choose appropriate instance types and sizes
- Implement caching strategies where appropriate
- Use CDNs for global distribution
- Optimize database performance
- Monitor and optimize application performance