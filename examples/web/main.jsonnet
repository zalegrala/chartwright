// Example chartwright entrypoint: one "web" component rendered as a Deployment +
// Service. Emits a Level-0 interchange document on stdout; feed it to the stamper:
//
//   stamp --jsonnet examples/web/main.jsonnet --out ./chart
//
local cw = import '../../lib/chartwright/chart.libsonnet';
local deployment = import '../../lib/chartwright/generators/deployment.libsonnet';
local service = import '../../lib/chartwright/generators/service.libsonnet';

cw.render(
  { name: 'acceptance', version: '0.1.0', appVersion: '2.6.0' },
  {
    web: {
      workload: 'Deployment',
      generators: ['deployment', 'service'],
      ports: [{ name: 'http', port: 3200 }],
      image: 'grafana/tempo:2.6.0',
      replicas: 1,
    },
  },
  { deployment: deployment, service: service },
)
