# Escalation and Ownership

Who owns what, for the cases that fall outside the CI Watcher's own remit (see `../ci-watcher/escalation-paths.md` for the CI-failure classification matrix - this doc covers component/infra ownership, not failure triage).

## CI job category ownership

When a CI failure needs to be routed to a person rather than a team-wide Slack ping, use the job's `ci-status-jobs.yaml` category:

| Job Category | Owner |
|---|---|
| ROSA CLI E2E / ROSA TF E2E | Amanda Katz (amakatz@redhat.com) |
| CAPA E2E | Mohamed ElSerngawy (melserng@redhat.com) |
| OCM FVT (HCP/Classic/GCP) | Jeff Frazier (jfrazier@redhat.com) |
| ROSA E2E STG / Conformance | Bo Meng (bmeng@redhat.com) |
| GAP E2E | Rohit Bhilare (rbhilare@redhat.com) |
| SRE Operator E2E | Dustin Row (drow@redhat.com) |

## SRE alert rule ownership layers

Separate from CI job ownership - this is for SRE-owned alert rules that fire on managed clusters, most control over least:

1. **managed-cluster-config (MCC)** - SRE-authored PrometheusRules deployed to clusters. We own `for:`, `expr`, labels directly. Most `*SRE`-suffixed alerts (e.g. `MachineHealthCheckUnterminatedShortCircuitSRE`, `ClusterMonitoringErrorBudgetBurnSRE`) live here.
2. **CAMO (configure-alertmanager-operator)** - alertmanager routing, inhibitions, silences, grouping. Use to tune *delivery* without touching the rule itself.
3. **RMO (route-monitor-operator)** - SLO burn-rate alerts for API/console probes (`api-ErrorBudgetBurn`, `console-ErrorBudgetBurn`, `api-RapidErrorBudgetBurn`).
4. **Upstream CMO (cluster-monitoring-operator)** - alert rules shipped by OpenShift itself. We can't change `for:`/`expr` here, only tune delivery via CAMO.

When investigating noisy or flapping alerts: check whether the rule is SRE-owned (MCC) before reaching for a CAMO workaround. Prefer fixing the source over routing around it.

**ROSA Rocket** (Jira team `[ROSA] Rocket`, ROSAENG project, board 12254) owns alert tuning work broadly - route alert-noise Jiras there, not to "SRE Operator E2E" or Operational Tooling.

## Build-farm / Prow infrastructure issues -> DPTP

Not every CI failure is a ROSA-side problem. Build-farm/Prow infrastructure issues - `initializing_namespace` races, build0N cluster auth outages, DNS/proxy failures reaching backplane, ci-operator namespace/lease contention - are owned by DPTP (Developer Productivity Test Platform), not any ROSA team.

**Don't just write these off as "transient, no action."** They recur, and without a tracked DPTP issue there's no owner and no way to distinguish a one-off blip from a systemic build-farm regression. When you hit one:

1. Ask `@chai-bot` in the relevant thread to search for an existing DPTP issue for the same failure signature, and open one if none exists.
2. If it's urgent (broad job impact, not just one PR), escalate directly in `#forum-ocp-testplatform` and ping `@dptp-triage`.
3. Note the DPTP ticket in whatever Jira you filed on the ROSA side so the two are cross-linked.

**Worked example (2026-09-25):** a stale Infoblox DNS entry for `squid.corp.redhat.com` was causing backplane-login timeouts across multiple build clusters, hitting `hive-e2e`, `pagerduty-operator`, and `ocm-fvt` jobs. chai-bot, when asked in the failure thread, immediately found two open ROSA-side Jiras (ROSAENG-67737/67738) and the broader DPTP tracker (DPTP-5138), and confirmed the failing test's exact proxy configuration from the step registry source. That's the pattern to follow: ask chai-bot first, it usually already has (or can build) the context before you escalate a human. The escalation to `#forum-ocp-testplatform` got only an automated "an engineer will respond in several hours" ack - DPTP response times for helpdesk pings are not fast, so file/link the Jira regardless of whether a human has replied yet.

## Compliance-alert style tickets

OHSS "Compliance Alert" tickets (e.g. SRE Cluster Admin Elevation alerts) are routine and high-volume; they don't usually need the same triage rigor as a CI failure. See `ohss-data-methodology`-style exclusions if you're ever reporting on incident counts - these are typically excluded because they're volatile and not representative of real incidents.
