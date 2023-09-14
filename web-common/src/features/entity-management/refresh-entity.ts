import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { createFileSaver } from "@rilldata/web-common/features/entity-management/file-actions";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  openFileUploadDialog,
  uploadFile,
} from "@rilldata/web-common/features/sources/modal/file-upload";
import { compileCreateSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import { overlay } from "@rilldata/web-common/layout/overlay-store";
import {
  createRuntimeServiceCreateTrigger,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export function createEntityRefresher() {
  const createTriggerMutation = createRuntimeServiceCreateTrigger();
  const fileSaver = createFileSaver();

  return async (source: V1Resource) => {
    const instanceId = get(runtime).instanceId;

    // non-local files can just be refreshed as is
    if (source.source.spec.sourceConnector !== "local_file") {
      return get(createTriggerMutation).mutateAsync({
        instanceId,
        data: {
          refreshTriggerSpec: {
            onlyNames: [source.meta.name],
          },
        },
      });
    }

    const sourceName = source.meta.name.name;
    const files = await openFileUploadDialog(false);
    if (!files.length) return Promise.reject();

    overlay.set({ title: `Importing ${sourceName}` });
    const filePath = await uploadFile(instanceId, files[0]);
    if (filePath === null) {
      return Promise.reject();
    }
    const yaml = compileCreateSourceYAML(
      {
        sourceName,
        path: filePath,
      },
      "local_file"
    );
    await fileSaver(
      getFilePathFromNameAndType(sourceName, EntityType.Table),
      yaml
    );
  };
}
