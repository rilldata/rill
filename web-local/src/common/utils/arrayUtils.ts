export function removeIfExists<T>(array: Array<T>, checker: (e: T) => boolean) {
  const index = array.findIndex(checker);
  if (index >= 0) {
    array.splice(index, 1);
  }
}

export function getMapFromArray<T>(
  array: Array<T>,
  keyGetter: (entity: T) => string | number
): Map<string | number, T> {
  const map = new Map();
  array.forEach((entity) => map.set(keyGetter(entity), entity));
  return map;
}
