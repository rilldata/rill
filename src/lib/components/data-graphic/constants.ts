export interface DomainCoordinates {
  x: (number | Date),
  y: (number | Date)
}
export const DEFAULT_COORDINATES = { x: undefined, y: undefined };

export const contexts = {
  config: 'rill:data-graphic:plot-config',
  scale(namespace: string) { return `rill:data-graphic:${namespace}-scale` },
  min(namespace: string) { return `rill:data-graphic:${namespace}-min` },
  max(namespace: string) { return `rill:data-graphic:${namespace}-max` },
}