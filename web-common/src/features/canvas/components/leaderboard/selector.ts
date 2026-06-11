import type { LeaderboardSpec } from "@rilldata/web-common/features/canvas/components/leaderboard";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function validateLeaderboardSchema(
  leaderboardSpec: LeaderboardSpec,
  metricsViewQuery: {
    metricsView: V1MetricsViewSpec | undefined;
    isLoading: boolean;
  },
): {
  isValid: boolean;
  error?: string;
  isLoading?: boolean;
} {
  if (metricsViewQuery.isLoading) {
    return {
      isValid: true,
      error: undefined,
      isLoading: true,
    };
  }
  const metricsView = metricsViewQuery.metricsView;
  if (!metricsView) {
    return {
      isValid: false,
      error: `Metrics view not found`,
    };
  }

  const allMeasures = metricsView?.measures?.map((m) => m.name as string) || [];
  const allDimensions =
    metricsView?.dimensions?.map((d) => d.name || (d.column as string)) || [];

  // Filter to only accessible fields, silently dropping any excluded by a
  // security policy.
  const measures = (leaderboardSpec?.measures || []).filter((m) =>
    allMeasures.includes(m),
  );
  const dimensions = (leaderboardSpec?.dimensions || []).filter((d) =>
    allDimensions.includes(d),
  );

  if (!measures.length || !dimensions.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }

  return {
    isValid: true,
    error: undefined,
  };
}
