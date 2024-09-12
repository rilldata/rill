import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export const FromURLParamTimeDimensionMap: Record<string, V1TimeGrain> = {
  "time.hour": V1TimeGrain.TIME_GRAIN_HOUR,
  "time.day": V1TimeGrain.TIME_GRAIN_DAY,
  "time.month": V1TimeGrain.TIME_GRAIN_MONTH,
};
export const ToURLParamTimeDimensionMap = reverseMap(
  FromURLParamTimeDimensionMap,
);

function reverseMap<K extends string | number, V extends string | number>(
  map: Partial<Record<K, V>>,
): Partial<Record<V, K>> {
  const revMap = {} as Partial<Record<V, K>>;
  for (const k in map) {
    revMap[map[k] as string | number] = map[k];
  }
  return revMap;
}
