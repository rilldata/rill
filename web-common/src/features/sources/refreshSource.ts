import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/modal/file-upload";
import { compileCreateSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import {
  runtimeServicePutFile,
  runtimeServiceTriggerRefresh,
} from "@rilldata/web-common/runtime-client";

export async function refreshSource(
  connector: string,
  filePath: string,
  sourceName: string,
  instanceId: string,
) {
  if (connector !== "local_file") {
    return runtimeServiceTriggerRefresh(instanceId, sourceName);
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
  if (dataFilePath === null) {
    return Promise.reject();
  }

  const yaml = compileCreateSourceYAML(
    {
      path: dataFilePath,
    },
    "local_file",
  );

  // Create source
  return runtimeServicePutFile(instanceId, filePath, {
    blob: yaml,
  });
}
