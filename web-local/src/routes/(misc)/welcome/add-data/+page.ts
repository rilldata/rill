import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { get } from "svelte/store";

export async function load() {
  // Fetch connectors and wait for the data to be loaded into cache
  await fetchAnalyzeConnectors(get(runtime).instanceId);
}
