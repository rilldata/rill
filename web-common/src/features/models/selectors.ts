import {
  createQueryServiceTableColumns,
  createRuntimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { TIMESTAMPS } from "../../lib/duckdb-data-types";

export function useModelFileIsEmpty(instanceId: string, modelName: string) {
  return createRuntimeServiceGetFile(instanceId, `/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}

export function useModelTimestampColumns(
  instanceId: string,
  modelName: string
) {
  return createQueryServiceTableColumns(
    instanceId,
    modelName,
    {},
    {
      query: {
        select: (data) =>
          data?.profileColumns
            .filter((c) => TIMESTAMPS.has(c.type))
            .map((c) => c.name),
      },
    }
  );
}
