# Roadmap

This document outlines near-term and aspirational goals for aws-ssm.

## 0.x Series Objectives
- Stabilize core EC2 session, list, interfaces, EKS features.
- Harden fuzzy finder performance for >5k instances.
- Improve least-privilege IAM examples.
- Add structured JSON output schemas and docs.
- Ship shell completions (bash/zsh/fish/powershell) via `aws-ssm completion`.
- Provide terminal recording of interactive usage (asciinema examples).

## Planned Features
- Plugin system (user-defined actions on selected instances).
- Metrics export (Prometheus HTTP endpoint) optional flag.
- Port forwarding multiplex improvements (auto-retry, health check).
- Session sharing / collaborative attach (multi-user).
- SSM document execution helper / automation integration.
- ASG & scaling workflows (already partial `cmd/asg.go`).
- EKS node group operations (describe / scale) expansion.

## Performance / UX
- Adaptive column widths and better truncation logic.
- Incremental instance loading (streaming) for huge fleets.
- Persistent favorites and tagging metadata file schema docs.

## Reliability
- Automatic exponential backoff and jitter everywhere AWS API is called.
- Circuit breaker for failing regions.

## Security
- Optional session MFA prompt before connect (policy enforcement doc).
- Audit logging output stream (JSON events: connect, command, forward start/stop).

## Long-Term Ideas
- Web/daemon mode exposing REST API for orchestration.
- TUI dashboard summarizing fleet health.
- Cross-account federation helper (role chooser interactive).

## Contributing to Roadmap
Open an issue with label `proposal` describing motivation, design sketch, and alternatives.

Roadmap items are not guarantees; priorities adapt to user feedback and security considerations.
