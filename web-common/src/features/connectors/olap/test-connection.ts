import { get } from "svelte/store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import { waitUntil } from "../../../lib/waitUtils";
import {
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
} from "../../../runtime-client";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { humanReadableErrorMessage } from "../../sources/modal/errors";

interface TestConnectorResult {
  success: boolean;
  error?: string;
}

export async function testOLAPConnector(
  instanceId: string,
  newConnectorFilePath: string,
  newConnectorName: string,
): Promise<TestConnectorResult> {
  // Wait a sec – give the runtime time to start reconciling the file
  await new Promise((resolve) => setTimeout(resolve, 1000));

  // Wait for the file to reconcile
  const fileArtifact = fileArtifacts.getFileArtifact(newConnectorFilePath);
  await waitUntil(() => !get(fileArtifact.reconciling), 500);

  // Check for errors
  const hasErrorsStore = fileArtifact.getHasErrors(queryClient, instanceId);
  const hasErrors = get(hasErrorsStore);
  if (hasErrors) {
    // Get the first error message
    const firstError = get(
      fileArtifact.getAllErrors(queryClient, instanceId),
    )[0].message;

    return {
      success: false,
      error: firstError,
    };
  }

  // Test the connection by calling `GetTables`
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
    return {
      success: false,
      error: humanReadableErrorMessage(
        newConnectorName,
        e?.response?.data?.code,
        e?.response?.data?.message,
      ),
    };
  }
}
