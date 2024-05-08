export function removeIfExists<T>(array: Array<T>, checker: (e: T) => boolean) {
  const index = array.findIndex(checker);
  if (index >= 0) {
    array.splice(index, 1);
    return true;
  }
  return false;
}

export function getMapFromArray<T, K>(
  array: T[],
  keyGetter: (entity: T) => K,
): Map<K, T> {
  const map = new Map<K, T>();
  for (const entity of array) {
    map.set(keyGetter(entity), entity);
  }
  return map;
}
