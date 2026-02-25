import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/modal/file-upload";
import { compileLocalFileSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  runtimeServiceCreateTrigger,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client/v2/gen";

export async function refreshSource(
  connector: string,
  filePath: string,
  sourceName: string,
  client: RuntimeClient,
) {
  if (connector !== "local_file") {
    return runtimeServiceCreateTrigger(client, {
      resources: [{ kind: ResourceKind.Source, name: sourceName }],
    });
  }

  // different logic for the file connector
  return replaceSourceWithUploadedFile(client, filePath);
}

export async function replaceSourceWithUploadedFile(
  client: RuntimeClient,
  filePath: string,
) {
  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  const dataFilePath = await uploadFile(client.instanceId, files[0]);
  if (dataFilePath === null || dataFilePath === undefined) {
    return Promise.reject();
  }

  const yaml = compileLocalFileSourceYAML(dataFilePath);

  // Create source
  return runtimeServicePutFile(client, {
    path: filePath,
    blob: yaml,
  });
}
