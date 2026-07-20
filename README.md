# chartwright

[![ci](https://github.com/zalegrala/chartwright/actions/workflows/ci.yml/badge.svg)](https://github.com/zalegrala/chartwright/actions/workflows/ci.yml)

> ⚠️ **Early work in progress.** APIs, formats, and structure are unstable and will break
> without notice.

**chartwright stamps a Helm chart out of your good deployment practices.**

You encode Kubernetes/application practice *once* — as small generator functions (probes,
security contexts, resource requests, PodDisruptionBudgets, config handling, …) — and every
component in every chart is stamped from them, uniformly. Learn a better practice? Change one
generator, re-stamp, and *every* chart inherits it. The Helm chart is disposable build output;
the generators are the institutional knowledge.

```
jsonnet (components + generators)  →  interchange JSON  →  stamper  →  Helm chart on disk
```

You describe components as data ("distributor is a Deployment with these ports and this
config"); generators shape them with your defaults; `helm.value()` marks the few things that
stay tunable at install time. A small, project-agnostic Go tool assembles the chart —
templates, `values.yaml`, `values.schema.json`, `Chart.yaml`. Output is byte-stable, so a
consumer's CI fails on uncommitted chart drift. First target is Grafana Tempo, but the core
knows nothing about Tempo — or even Kubernetes semantics.

## Why not just write the chart yourself?

Because a hand-written chart is a *liability*, and generators are an *asset*:

- **Practices are encoded once and applied everywhere.** The `nginx` in
  [`examples/minimal`](./examples/minimal/main.jsonnet) — ~15 lines that mention none of this —
  stamps out a Deployment with a hardened pod + container `securityContext`, resource requests,
  and readiness/liveness probes, because the deployment generator bakes them in. Every
  component gets the same good defaults for free.
- **Improvements compound centrally.** Adopt a better probe shape or a new hardening default in
  [`workload.libsonnet`](./lib/chartwright/workload.libsonnet), re-stamp, and every chart moves
  forward at once — instead of drifting apart as each hand-edited chart reinvents the wheel and
  bad defaults propagate.
- **Consequences are reviewable.** The rendered chart is a committed artifact, so a jsonnet
  change shows the exact Kubernetes diff (new mount, changed probe, extra RBAC) in review.
- **One structured way to pass config** — the whole app config as a single opaque value Tempo
  validates at runtime, not a 1:1 mapping of every knob (tempo-distributed's maintenance tax).

The thesis in one line: **users want to *consume* Helm; authors don't want to *author* it** —
so author your practices in jsonnet and let the machine produce the Helm. See
[`DESIGN.md`](./DESIGN.md) for the full rationale.

## Status

| Component | Status |
|-----------|--------|
| Stamper core (interchange → chart) | ✅ working |
| Hole-marker lowering pass | ✅ working |
| Jsonnet authoring layer (`helm.value`, generators) | ✅ working (deployment/service/statefulset/pdb/configmap/vpa/servicemonitor) |
| Config-mount primitive (structured config → ConfigMap → mount) | ✅ working |
| CRD generators + kubeconform CRD validation | ✅ working |
| Version/capability gating (`.Capabilities`, kubeVersion) | ✅ working |
| Tempo descriptors + example wiring | ⏳ not started |

See [`DESIGN.md` §14](./DESIGN.md) for the roadmap.

## Try it

```bash
# Smallest example — one component → Deployment + Service (start here):
go run ./cmd/stamp --jsonnet examples/minimal/main.jsonnet --out /tmp/chart

# Fuller showcase — config-mount, CRDs, capability gates, chart-scoped RBAC:
go run ./cmd/stamp --jsonnet examples/web/main.jsonnet --out /tmp/chart

# Version/capability gating — apiVersion switch + whole-resource gates by k8s version/API:
go run ./cmd/stamp --jsonnet examples/version-gating/main.jsonnet --out /tmp/chart

# Tempo-flavored demo — 4 microservices components + one structured config mounted across all:
go run ./cmd/stamp --jsonnet examples/tempo/main.jsonnet --out /tmp/chart

# Or from a hand-written interchange JSON document (no jsonnet):
go run ./cmd/stamp --in testdata/installable.json --out /tmp/chart
```

`examples/minimal/main.jsonnet` is ~15 readable lines and the best first read. Renders into an
installable Helm chart under `/tmp/chart`. `--check` compares against an existing chart and
exits non-zero on drift (for CI); `--jsonnet <file>` runs jsonnet and uses its stdout as the
interchange input.

## How this is being built

This project is developed openly and largely with AI assistance (Claude Code). The design
conversations, specs, and step-by-step implementation plans are committed in-repo under
[`DESIGN.md`](./DESIGN.md) and [`docs/`](./docs/) rather than hidden — the process is part of
the artifact. Feedback welcome.

## License

[Apache License 2.0](./LICENSE).
