import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
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
  sourceName: string,
  instanceId: string,
) {
  if (connector !== "local_file") {
    return runtimeServiceTriggerRefresh(instanceId, sourceName);
  }

  // different logic for the file connector

  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  const filePath = await uploadFile(instanceId, files[0]);
  if (filePath === null) {
    return Promise.reject();
  }
  const yaml = compileCreateSourceYAML(
    {
      sourceName,
      path: filePath,
    },
    "local_file",
  );
  return runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(sourceName, EntityType.Table),
    {
      blob: yaml,
    },
  );
}

export async function replaceSourceWithUploadedFile(
  instanceId: string,
  sourceName: string,
) {
  const artifactPath = getFileAPIPathFromNameAndType(
    sourceName,
    EntityType.Table,
  );

  const files = await openFileUploadDialog(false);
  if (!files.length) return Promise.reject();

  const filePath = await uploadFile(instanceId, files[0]);
  if (filePath === null) {
    return Promise.reject();
  }

  const yaml = compileCreateSourceYAML(
    {
      sourceName,
      path: filePath,
    },
    "local_file",
  );

  // Create source
  const resp = await runtimeServicePutFile(instanceId, artifactPath, {
    blob: yaml,
  });

  return resp;
}
