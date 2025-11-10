# Security Policy

## Supported Versions
This project is pre-1.0. Expect rapid iteration. We aim to support the latest minor Go releases (currently 1.23 and 1.24) and the latest two tagged versions of this project.

## Reporting a Vulnerability
Please DO NOT open a public issue for security concerns.

1. Create a private GitHub Security Advisory ("Report a vulnerability" from the repo's Security tab), or
2. Email: security+aws-ssm@johnlam.dev (PGP available on keyservers: 0xA1B2C3D4E5F6A7B8)

Provide:
- Description and impact
- Steps to reproduce / proof-of-concept
- Affected versions
- Suggested remediation (if available)

You will receive acknowledgment within 72 hours. We will coordinate a fix and disclosure timeline (typically <= 30 days depending on severity).

## Disclosure
We prefer coordinated disclosure. After a fix is released, an advisory will be published with CVE (if applicable).

## Security Principles
- No secrets are stored locally beyond AWS SDK default credential chain.
- Sessions tunnel via AWS SSM; no inbound ports opened.
- Dependencies scanned via CI (Trivy, Gosec, SBOM generation).
- Avoid hardcoded credentials (CI enforces). 

## Hardening Recommendations
- Use least-privilege IAM policies (scope EC2 + EKS resources where possible).
- Rotate credentials and enforce MFA for AWS profiles used.
- Audit session usage via AWS CloudTrail (StartSession / TerminateSession / SendCommand).

## Hall of Fame
Contributors responsibly disclosing security issues may be recognized in release notes.

Thank you for helping keep the project secure.
