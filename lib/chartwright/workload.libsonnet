// workload.libsonnet — shared pieces for pod-bearing workloads (Deployment,
// StatefulSet). This is where good deployment PRACTICE lives, encoded once and
// applied to every component: hardened security contexts, resource requests,
// health probes, and config wiring. Improve a default here → every component in
// every chart re-stamps with it. That is the point of chartwright: the chart is
// throwaway output; this file is the institutional knowledge. (DESIGN §1)
local helm = import 'helm.libsonnet';
local mounts = import 'mounts.libsonnet';

{
  labels(c):: { 'app.kubernetes.io/name': c.name },
  selector(c):: { matchLabels: $.labels(c) },

  // podTemplate builds the shared spec.template for a component, baking in
  // sensible, overridable defaults. Everything data-driven off the descriptor.
  podTemplate(c):: {
    local ports = std.get(c, 'ports', []),
    local hasConfigs = std.length(std.get(c, 'configs', [])) > 0,
    local annotations = mounts.checksumAnnotations(c),
    // health probe target: explicit `health`, else the first port, else none.
    local health = std.get(c, 'health', {
      path: '/',
      port: if std.length(ports) > 0 then ports[0].name else null,
    }),
    local probe = { httpGet: { path: health.path, port: health.port } },

    metadata: {
      labels: $.labels(c),
      [if std.length(annotations) > 0 then 'annotations']: annotations,
    },
    spec: {
      // GOOD DEFAULT: run as non-root with the runtime's default seccomp profile.
      securityContext: {
        runAsNonRoot: true,
        seccompProfile: { type: 'RuntimeDefault' },
      },
      containers: [{
        name: c.name,
        image: helm.value(c.name + '.image', std.get(c, 'image', 'busybox:latest'),
                          { render: 'quote' }),
        [if std.length(ports) > 0 then 'ports']:
          [{ name: p.name, containerPort: p.port } for p in ports],
        // GOOD DEFAULT: resource requests set (tunable as one block value).
        resources: helm.blockValue(c.name + '.resources',
                                   std.get(c, 'resources', { requests: { cpu: '100m', memory: '128Mi' } })),
        // GOOD DEFAULT: a hardened container security context.
        securityContext: {
          allowPrivilegeEscalation: false,
          readOnlyRootFilesystem: true,
          capabilities: { drop: ['ALL'] },
        },
        // GOOD DEFAULT: readiness + liveness probes when there's a port to probe.
        [if health.port != null then 'readinessProbe']: probe,
        [if health.port != null then 'livenessProbe']: probe,
        [if hasConfigs then 'volumeMounts']: mounts.mounts(c),
      }],
      [if hasConfigs then 'volumes']: mounts.volumes(c),
    },
  },
}
