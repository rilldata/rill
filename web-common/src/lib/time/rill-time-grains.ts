/**
 * Functions that depend on RillTime types.
 * Separated from new-grains.ts to avoid circular dependency with RillTime.ts
 */
import {
  RillLegacyDaxInterval,
  RillLegacyIsoInterval,
  type RillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  GrainAliasToV1TimeGrain,
  getSmallestGrain,
  type TimeGrainAlias,
} from "./new-grains";

export function getRangePrecision(rillTime: RillTime) {
  const asOfSnap = rillTime.asOfLabel?.snap;

  const asOfSnapV1Grain = GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];
  const rangeV1Grain = rillTime.rangeGrain;
  const intervalV1Grain = rillTime.interval.getGrain();

  return getSmallestGrain([asOfSnapV1Grain, rangeV1Grain, intervalV1Grain]);
}

export function getAggregationGrain(rillTime: RillTime | undefined) {
  if (!rillTime) return undefined;

  const asOfSnap = rillTime.asOfLabel?.snap;

  const asOfSnapV1Grain = GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];
  const rangeV1Grain = rillTime.rangeGrain;
  const intervalV1Grain = rillTime.interval.getGrain();

  return getSmallestGrain([asOfSnapV1Grain, rangeV1Grain, intervalV1Grain]);
}

export function getTruncationGrain(rillTime: RillTime | undefined) {
  if (!rillTime) return undefined;

  const asOfSnap = rillTime.asOfLabel?.snap;

  if (asOfSnap) return GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];

  if (rillTime.interval instanceof RillLegacyIsoInterval) {
    return rillTime.interval.getGrain();
  }

  if (rillTime.interval instanceof RillLegacyDaxInterval) {
    if (rillTime.interval.name.endsWith("C")) return undefined;
    return V1TimeGrain.TIME_GRAIN_DAY;
  }

  return undefined;
}
