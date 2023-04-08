import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";

export function modelIsEmpty(instanceId, modelName) {
  return createRuntimeServiceGetFile(instanceId, `/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}
