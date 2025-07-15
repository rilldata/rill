import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
  connectorServiceListDatabaseSchemas,
  getConnectorServiceListDatabaseSchemasQueryKey,
  connectorServiceListTables,
  getConnectorServiceListTablesQueryKey,
  connectorServiceGetTable,
  getConnectorServiceGetTableQueryKey,
} from "../../../runtime-client";
import { humanReadableErrorMessage } from "../../sources/errors/errors";

interface TestConnectorResult {
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

export async function testListDatabaseSchemas(instanceId: string, params: any) {
  const queryKey = getConnectorServiceListDatabaseSchemasQueryKey({
    instanceId,
    ...params,
  });
  const queryFn = () =>
    connectorServiceListDatabaseSchemas({
      instanceId,
      ...params,
    });

  try {
    await queryClient.fetchQuery({ queryKey, queryFn });
    return { success: true };
  } catch (e) {
    const originalMessage = e?.response?.data?.message;
    return {
      success: false,
      error: humanReadableErrorMessage(
        params?.source || "source",
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

export async function testListTables(instanceId: string, params: any) {
  const queryKey = getConnectorServiceListTablesQueryKey({
    instanceId,
    ...params,
  });
  const queryFn = () =>
    connectorServiceListTables({
      instanceId,
      ...params,
    });

  try {
    await queryClient.fetchQuery({ queryKey, queryFn });
    return { success: true };
  } catch (e) {
    const originalMessage = e?.response?.data?.message;
    return {
      success: false,
      error: humanReadableErrorMessage(
        params?.source || "source",
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

export async function testGetTable(instanceId: string, params: any) {
  const queryKey = getConnectorServiceGetTableQueryKey({
    instanceId,
    ...params,
  });
  const queryFn = () =>
    connectorServiceGetTable({
      instanceId,
      ...params,
    });

  try {
    await queryClient.fetchQuery({ queryKey, queryFn });
    return { success: true };
  } catch (e) {
    const originalMessage = e?.response?.data?.message;
    return {
      success: false,
      error: humanReadableErrorMessage(
        params?.source || "source",
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}
