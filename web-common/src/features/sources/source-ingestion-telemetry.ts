import type { EntityActionInstance } from "@rilldata/web-common/features/entity-management/entity-action-queue";
import {
  emitSourceErrorTelemetry,
  emitSourceSuccessTelemetry,
} from "@rilldata/web-common/features/sources/sourceUtils";
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
      actionInstance.telemetryParams.space,
      actionInstance.telemetryParams.screenName,
      resource.meta.reconcileError,
      connectorToSourceConnectionType[connectorName],
      resource.source.spec.properties?.path ?? ""
    );
  } else {
    // Success
    emitSourceSuccessTelemetry(
      actionInstance.telemetryParams.space,
      actionInstance.telemetryParams.screenName,
      actionInstance.telemetryParams.medium,
      connectorToSourceConnectionType[connectorName],
      resource.source.spec.properties?.path ?? ""
    );
  }
}
