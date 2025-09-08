import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
  connectorServiceListDatabaseSchemas,
  getConnectorServiceListDatabaseSchemasQueryKey,
} from "../../../runtime-client";
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

// Test the connection for non-OLAP connectors by calling `ListDatabaseSchemas`
export async function testNonOLAPConnector(
  instanceId: string,
  newConnectorName: string,
): Promise<TestConnectorResult> {
  const queryKey = getConnectorServiceListDatabaseSchemasQueryKey({
    instanceId,
    connector: newConnectorName,
  });
  const queryFn = () =>
    connectorServiceListDatabaseSchemas({
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
