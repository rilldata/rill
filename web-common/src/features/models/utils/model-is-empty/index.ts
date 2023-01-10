import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";

export function modelIsEmpty(instanceId, modelName) {
  return useRuntimeServiceGetFile(instanceId, `/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}
