import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export function getOrgAndProjectForEmbed(runtimeClient: RuntimeClient) {
  return derived(createRuntimeServiceGetInstance(runtimeClient, {}), (resp) => {
    const urlParts = resp.data?.instance?.frontendUrl?.split("/") ?? [];
    const organization = urlParts[urlParts.length - 2] ?? "";
    const project = urlParts[urlParts.length - 1] ?? "";
    return { organization, project };
  });
}
