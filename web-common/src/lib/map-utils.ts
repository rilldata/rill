export function reverseMap<
  K extends string | number,
  V extends string | number,
>(map: Partial<Record<K, V>>): Partial<Record<V, K>> {
  const revMap = {} as Partial<Record<V, K>>;
  for (const k in map) {
    revMap[map[k] as string | number] = k;
  }
  return revMap;
}
