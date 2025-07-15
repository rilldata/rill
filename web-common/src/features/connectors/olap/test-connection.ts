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
  type ConnectorServiceListDatabaseSchemasParams,
  type ConnectorServiceListTablesParams,
  type ConnectorServiceGetTableParams,
} from "../../../runtime-client";
import { humanReadableErrorMessage } from "../../sources/errors/errors";

interface TestConnectorResult {
  success: boolean;
  error?: string;
  details?: string;
}

export async function testOLAPConnector(
  instanceId: string,
  params: { connector: string },
): Promise<TestConnectorResult> {
  // Test the connection by calling `ListTables`
  const queryKey = getConnectorServiceOLAPListTablesQueryKey({
    instanceId,
    ...params,
  });
  const queryFn = () =>
    connectorServiceOLAPListTables({
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
        params.connector,
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

export async function testListDatabaseSchemas(
  instanceId: string,
  params: ConnectorServiceListDatabaseSchemasParams,
) {
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
        params?.connector,
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

export async function testListTables(
  instanceId: string,
  params: ConnectorServiceListTablesParams,
) {
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
        params?.connector,
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}

export async function testGetTable(
  instanceId: string,
  params: ConnectorServiceGetTableParams,
) {
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
        params?.connector,
        e?.response?.data?.code,
        originalMessage,
      ),
      details: originalMessage,
    };
  }
}
