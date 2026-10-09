import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import { getAdaptiveTimeGrain } from "@rilldata/web-common/lib/time/grains";
import {
  V1TimeGrain,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";

/**
 * Sets `selectedTimeRange.interval` to the adaptive grain for the selected time range.
 * The grain is derived from the range's start and end once they are resolved;
 * before that only the rill-time precision is available.
 * Consumers such as export and measure validation read the grain straight from the explore state,
 * so it is stored concretely even in adaptive mode.
 */
export function applyAdaptiveTimeGrain(
  exploreState: Partial<ExploreState>,
  metricsViewSpec: V1MetricsViewSpec,
) {
  const timeRange = exploreState.selectedTimeRange;
  if (!exploreState.adaptiveTimeGrain || !timeRange) return;

  let interval: Interval | undefined;
  if (timeRange.start && timeRange.end) {
    const zone = exploreState.selectedTimezone;
    interval = Interval.fromDateTimes(
      DateTime.fromJSDate(timeRange.start, { zone }),
      DateTime.fromJSDate(timeRange.end, { zone }),
    );
  }

  let parsed: RillTime | undefined;
  if (timeRange.name) {
    try {
      parsed = parseRillTime(timeRange.name);
    } catch {
      // Custom ranges and legacy presets are not rill-time expressions.
    }
  }

  const grain = getAdaptiveTimeGrain(
    interval,
    metricsViewSpec.smallestTimeGrain || V1TimeGrain.TIME_GRAIN_MINUTE,
    parsed,
  );
  if (grain) timeRange.interval = grain;
}
