import {
  runtimeServiceGetFile,
  getRuntimeServiceGetFileQueryKey,
} from "@rilldata/web-common/runtime-client/index.js";
import type { QueryFunction } from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error } from "@sveltejs/kit";

type FileType =
  | "md"
  | "sql"
  | "yaml"
  | "yml"
  | "txt"
  | "csv"
  | "parquet"
  | "json";

export async function load({ parent, url }) {
  const { instanceId } = await parent();
  const path = url.pathname.slice(5);
  const fileName = path.split("/").pop();

  if (!fileName || !path) throw error(404, "Not found");

  let name: string;
  let type: FileType = "txt";

  if (fileName.startsWith(".")) {
    name = fileName;
  } else {
    const parts = fileName.split(".");
    type = parts.pop() as FileType;

    if (parts.length > 1) {
      name = parts.join(".");
    } else {
      name = parts[0];
    }
  }

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetFile>>
  > = ({ signal }) => runtimeServiceGetFile(instanceId, path, signal);

  const fileQuery = queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(instanceId, path),
    queryFn,
  });

  try {
    const file = await fileQuery;

    return {
      file: {
        ...file,
        path,
        name,
        type,
      },
    };
  } catch (e) {
    throw error(404, "Not found");
  }
}
