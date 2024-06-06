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

export function createBatches<T>(array: T[], batchSize: number): T[][] {
  const batches: T[][] = [];
  for (let i = 0; i < array.length; i += batchSize) {
    batches.push(array.slice(i, i + batchSize));
  }
  return batches;
}
