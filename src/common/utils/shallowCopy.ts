export function shallowCopy<Entity>(source: Entity, target: Entity): void {
  Object.keys(source).forEach((k) => {
    target[k] = source[k];
  });
}
