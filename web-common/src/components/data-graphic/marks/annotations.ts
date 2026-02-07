import type { V1MetricsViewAnnotationsResponseAnnotation } from "@rilldata/web-common/runtime-client";
import type { DateTime } from "luxon";

export type Annotation = V1MetricsViewAnnotationsResponseAnnotation & {
  startTime: DateTime;
  endTime?: DateTime;
  formattedTimeOrRange: string;
};
