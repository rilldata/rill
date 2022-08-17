export function getMapFromArray<T>(
  array: Array<T>,
  keyGetter: (entity: T) => string | number
): Map<string | number, T> {
  const map = new Map();
  array.forEach((entity) => map.set(keyGetter(entity), entity));
  return map;
}
