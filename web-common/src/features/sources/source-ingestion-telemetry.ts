import type { EntityActionInstance } from "@rilldata/web-common/features/entity-management/entity-action-queue";
import {
  emitSourceErrorTelemetry,
  emitSourceSuccessTelemetry,
} from "@rilldata/web-common/features/sources/sourceUtils";
import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
import { connectorToSourceConnectionType } from "@rilldata/web-common/metrics/service/SourceEventTypes";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

export function sourceIngestionTelemetry(
  resource: V1Resource,
  actionInstance: EntityActionInstance
) {
  const connectorName = resource.source.spec.sourceConnector;
  if (resource.meta.reconcileError) {
    // Error
    emitSourceErrorTelemetry(
      MetricsEventSpace.Modal,
      actionInstance.screenName,
      resource.meta.reconcileError,
      connectorToSourceConnectionType[connectorName],
      resource.source.spec.properties?.path ?? ""
    );
  } else {
    // Success
    emitSourceSuccessTelemetry(
      MetricsEventSpace.Modal,
      actionInstance.screenName,
      actionInstance.behaviourEventMedium,
      connectorToSourceConnectionType[connectorName],
      resource.source.spec.properties?.path ?? ""
    );
  }
}
