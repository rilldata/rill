import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  runtimeServicePutFile,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { removeLeadingSlash } from "../../entity-management/entity-mappers";
import { invalidate } from "$app/navigation";

export async function saveFile(
  queryClient: QueryClient,
  path: string,
  blob: string,
) {
  const instanceId = get(runtime).instanceId;

  await runtimeServicePutFile(instanceId, path, {
    blob,
  });

  // Invalidate `GetFile` query
  // await queryClient.invalidateQueries(
  //   getRuntimeServiceGetFileQueryKey(instanceId, removeLeadingSlash(path)),
  // );

  queryClient
    .getQueryCache()
    .find(
      getRuntimeServiceGetFileQueryKey(instanceId, removeLeadingSlash(path)),
    )
    ?.setState((old) => {
      return {
        ...old,
        data: {
          ...old.data,
          blob,
        },
      };
    });

  await invalidate(removeLeadingSlash(path));
  // If it's a rill.yaml file, invalidate the dev JWT queries
  if (path === "rill.yaml") {
    await queryClient.invalidateQueries(getRuntimeServiceIssueDevJWTQueryKey());
  }
}
