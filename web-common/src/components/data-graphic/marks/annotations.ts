import type { V1MetricsViewAnnotationsResponseAnnotation } from "@rilldata/web-common/runtime-client";

export type Annotation = V1MetricsViewAnnotationsResponseAnnotation & {
  startTime: Date;
  endTime?: Date;
  formattedTimeOrRange: string;
};
