import type { LeaderboardSpec } from "@rilldata/web-common/features/canvas/components/leaderboard";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function validateLeaderboardSchema(
  leaderboardSpec: LeaderboardSpec,
  metricsView: V1MetricsViewSpec | undefined,
): {
  isValid: boolean;
  error?: string;
} {
  if (!metricsView) {
    return {
      isValid: false,
      error: `Metrics view not found`,
    };
  }

  const allMeasures = metricsView?.measures?.map((m) => m.name as string) || [];
  const allDimensions =
    metricsView?.dimensions?.map((d) => d.name || (d.column as string)) || [];

  let measures = leaderboardSpec?.measures || [];
  let dimensions = leaderboardSpec?.dimensions || [];

  if (!measures.length || !dimensions.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }

  measures = measures.filter((c) => allMeasures.includes(c));
  dimensions = dimensions.filter((c) => allDimensions.includes(c));

  const validateMeasuresRes = validateMeasures(metricsView, measures);
  if (!validateMeasuresRes.isValid) {
    const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
    return {
      isValid: false,
      error: `Invalid measure(s) "${invalidMeasures}" selected for the table`,
    };
  }

  const validateDimensionsRes = validateDimensions(metricsView, dimensions);

  if (!validateDimensionsRes.isValid) {
    const invalidDimensions =
      validateDimensionsRes.invalidDimensions.join(", ");

    return {
      isValid: false,
      error: `Invalid dimension(s) "${invalidDimensions}" selected for the table`,
    };
  }
  return {
    isValid: true,
    error: undefined,
  };
}
