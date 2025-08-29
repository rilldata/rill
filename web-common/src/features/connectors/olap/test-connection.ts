import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
} from "../../../runtime-client";
import { humanReadableErrorMessage } from "../../sources/errors/errors";

export interface TestConnectorResult {
  success: boolean;
  error?: string;
  details?: string;
}

export async function testOLAPConnector(
  instanceId: string,
  newConnectorName: string,
): Promise<TestConnectorResult> {
  // Test the connection by calling `ListTables`

  const queryKey = getConnectorServiceOLAPListTablesQueryKey({
    instanceId,
    connector: newConnectorName,
  });
  const queryFn = () =>
    connectorServiceOLAPListTables({
      instanceId,
      connector: newConnectorName,
    });

  try {
    await queryClient.fetchQuery({ queryKey, queryFn });
    return { success: true };
  } catch (e) {
    const originalMessage = e?.response?.data?.message;
    return {
      success: false,
      error: humanReadableErrorMessage(
        newConnectorName,
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

// Poll the connector resource status using `runtimeServiceGetResource`
// to check if the connector configuration is valid and reconciled
export async function testNonOlapConnector(
  instanceId: string,
  newConnectorName: string,
): Promise<TestConnectorResult> {
  // Test the connection by polling the connector resource status
  const maxAttempts = 10; // Maximum number of polling attempts
  const pollInterval = 2000; // 2 seconds between attempts
  const maxWaitTime = maxAttempts * pollInterval; // Maximum total wait time

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const resource = await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.kind": "connector",
          "name.name": newConnectorName,
        }),
        queryFn: () =>
          runtimeServiceGetResource(instanceId, {
            "name.kind": "connector",
            "name.name": newConnectorName,
          }),
      });

      // Check if resource has reconcile errors
      if (resource.resource?.meta?.reconcileError) {
        return {
          success: false,
          error: "Connector configuration failed to reconcile",
          details: resource.resource.meta.reconcileError,
        };
      }

      // Check if resource is healthy (reconcile status is idle)
      if (
        resource.resource?.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE"
      ) {
        return { success: true };
      }

      // If this is the last attempt, return the current status
      if (attempt === maxAttempts) {
        return {
          success: false,
          error: "Connector reconciliation timeout",
          details: `Connector is still reconciling after ${maxWaitTime / 1000} seconds. Current status: ${resource.resource?.meta?.reconcileStatus || "unknown"}`,
        };
      }

      // Wait before the next poll attempt
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    } catch (error) {
      // If this is the last attempt, return the error
      if (attempt === maxAttempts) {
        return {
          success: false,
          error: "Failed to check connector status",
          details:
            error?.message ||
            "Unknown error occurred while polling connector status",
        };
      }

      // Wait before retrying on error
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }
  }

  // This should never be reached, but just in case
  return {
    success: false,
    error: "Unexpected error during connector testing",
    details: "An unexpected error occurred while testing the connector",
  };
}
