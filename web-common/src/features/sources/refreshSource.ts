import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/modal/file-upload";
import { compileLocalFileModelYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import {
  runtimeServiceCreateTrigger,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";

export async function refreshSource(
  connector: string,
  filePath: string,
  sourceName: string,
  instanceId: string,
) {
  if (connector !== "local_file") {
    return runtimeServiceCreateTrigger(instanceId, {
      resources: [{ kind: ResourceKind.Model, name: sourceName }],
    });
  }

  // different logic for the file connector
  return replaceSourceWithUploadedFile(instanceId, filePath);
}

export async function replaceSourceWithUploadedFile(
  instanceId: string,
  filePath: string,
) {
  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  const dataFilePath = await uploadFile(instanceId, files[0]);
  if (dataFilePath === null || dataFilePath === undefined) {
    return Promise.reject();
  }

  const yaml = compileLocalFileModelYAML(dataFilePath);

  // Create model instead of source
  return runtimeServicePutFile(instanceId, {
    path: filePath,
    blob: yaml,
  });
}
