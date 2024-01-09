import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  runtimeServicePutFile,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";

export async function saveFile(
  queryClient: QueryClient,
  path: string,
  blob: string,
) {
  const instanceId = get(runtime).instanceId;

  await runtimeServicePutFile(instanceId, path, {
    blob: blob,
  });

  // Invalidate `GetFile` query
  queryClient.invalidateQueries(
    getRuntimeServiceGetFileQueryKey(instanceId, path),
  );

  // If it's a rill.yaml file, invalidate the dev JWT queries
  if (path === "rill.yaml") {
    queryClient.invalidateQueries(getRuntimeServiceIssueDevJWTQueryKey());
  }
}
