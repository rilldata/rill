import { categorizeSourceError } from "@rilldata/web-common/features/sources/modal/errors";
import { getFileTypeFromPath } from "@rilldata/web-common/features/sources/sourceUtils";
import {
  behaviourEvent,
  errorEventHandler,
} from "@rilldata/web-common/metrics/initMetrics";
import type { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import type {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { SourceConnectionType } from "@rilldata/web-common/metrics/service/SourceEventTypes";

export function emitSourceErrorTelemetry(
  space: MetricsEventSpace,
  screenName: MetricsEventScreenName,
  errorMessage: string,
  connectionType: SourceConnectionType,
  fileName: string,
) {
  const categorizedError = categorizeSourceError(errorMessage);
  const fileType = getFileTypeFromPath(fileName);
  const isGlob = fileName.includes("*");

  errorEventHandler?.fireSourceErrorEvent(
    space,
    screenName,
    categorizedError,
    connectionType,
    fileType,
    isGlob,
  );
}

export function emitSourceSuccessTelemetry(
  space: MetricsEventSpace,
  screenName: MetricsEventScreenName,
  medium: BehaviourEventMedium,
  connectionType: SourceConnectionType,
  fileName: string,
) {
  const fileType = getFileTypeFromPath(fileName);
  const isGlob = fileName.includes("*");

  behaviourEvent?.fireSourceSuccessEvent(
    medium,
    screenName,
    space,
    connectionType,
    fileType,
    isGlob,
  );
}
