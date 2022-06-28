export const contexts = {
  config: 'rill:data-graphic:plot-config',
  scale(namespace) { return `rill:data-graphic:${namespace}-scale` },
  min(namespace) { return `rill:data-graphic:${namespace}-min` },
  max(namespace) { return `rill:data-graphic:${namespace}-max` },
}