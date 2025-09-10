import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
  runtimeServiceGetResource,
} from "../../../runtime-client";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { humanReadableErrorMessage } from "../../sources/errors/errors";

export interface TestConnectorResult {
  success: boolean;
  error?: string;
  details?: string;
}

// Test the connection by calling  `ListTables`
export async function testOLAPConnector(
  instanceId: string,
  newConnectorName: string,
): Promise<TestConnectorResult> {
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

// Poll the connector resource status to check reconcile status
export async function pollConnectorReconcileStatus(
  instanceId: string,
  connectorName: string,
): Promise<TestConnectorResult> {
  const maxAttempts = 15; // 30 seconds total (2s * 15)
  const pollInterval = 2000; // 2 seconds
  const maxWaitTime = maxAttempts * pollInterval;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const resource = await runtimeServiceGetResource(instanceId, {
        "name.kind": ResourceKind.Connector,
        "name.name": connectorName,
      });

      // Check if there's a reconcile error
      if (resource.resource?.meta?.reconcileError) {
        return {
          success: false,
          error: "Connector configuration failed to reconcile",
          details: resource.resource.meta.reconcileError,
        };
      }

      // Check the reconcile status
      const reconcileStatus = resource.resource?.meta?.reconcileStatus;

      if (reconcileStatus === "RECONCILE_STATUS_IDLE") {
        return { success: true };
      }

      // Still reconciling, continue polling
      if (attempt < maxAttempts) {
        await new Promise((resolve) => setTimeout(resolve, pollInterval));
        continue;
      }

      // Last attempt and still not idle
      return {
        success: false,
        error: "Connector reconciliation timeout",
        details: `Connector is still reconciling after ${maxWaitTime / 1000} seconds. Current status: ${reconcileStatus || "unknown"}`,
      };
    } catch (error) {
      // Resource not found is expected initially after file creation
      if (error?.status === 404 || error?.response?.status === 404) {
        if (attempt === maxAttempts) {
          return {
            success: false,
            error: "Connector resource was never created",
            details: `The connector "${connectorName}" file was created but the runtime never processed it into a resource. This may indicate a configuration error or runtime issue.`,
          };
        }

        // Wait and try again
        await new Promise((resolve) => setTimeout(resolve, pollInterval));
        continue;
      }

      // For the last attempt, provide detailed error info
      if (attempt === maxAttempts) {
        return {
          success: false,
          error: "Failed to check connector status",
          details:
            error?.response?.data?.message ||
            error?.message ||
            "Unknown error occurred",
        };
      }
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }
  }

  // Fallback (should never reach here)
  return {
    success: false,
    error: "Unexpected error during connector testing",
    details: "An unexpected error occurred while testing the connector",
  };
}
