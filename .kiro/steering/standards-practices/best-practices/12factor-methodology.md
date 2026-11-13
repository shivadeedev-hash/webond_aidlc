---
inclusion: always
---

# Twelve-Factor Methodology

The twelve-factor methodology helps you build modern, scalable, and maintainable software-as-a-service apps. The methodology is technology agnostic and has become a widely-adopted approach to developing cloud-native applications.

These principles were captured in the following 12 factors:

- I. Codebase
One codebase tracked in revision control, many deploys
- II. Dependencies
Explicitly declare and isolate dependencies
- III. Config
Store config in the environment
- IV. Backing services
Treat backing services as attached resources
- V. Build, release, run
Strictly separate build and run stages
- VI. Processes
Execute the app as one or more stateless processes
- VII. Port binding
Export services via port binding
- VIII. Concurrency
Scale out via the process model
- IX. Disposability
Maximize robustness with fast startup and graceful shutdown
- X. Dev/prod parity
Keep development, staging, and production as similar as possible
- XI. Logs
Treat logs as event streams
- XII. Admin processes
Run admin/management tasks as one-off processes

Reference:
https://12factor.net/
https://github.com/twelve-factor/twelve-factor

The six pillars of the AWS Well-Architected Framework
Let’s explore the six pillars of the AWS Well-Architected Framework, what each aims to achieve, and where the twelve-factors concepts intersect with AWS guidance.

The following figures shows the twelve factors and how they map to processes in AWS, which are described in this section.

![alt text](https://d2908q01vomqb2.cloudfront.net/fc074d501302eb2b93e2554793fcaf50b3bf7291/2025/02/11/image-1.png)