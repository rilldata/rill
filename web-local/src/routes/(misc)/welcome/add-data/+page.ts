import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
import { getLocalRuntimeClient } from "@rilldata/web-local/lib/runtime-client.ts";

export async function load() {
  const client = getLocalRuntimeClient();

  // Fetch connectors and wait for the data to be loaded into cache
  await fetchAnalyzeConnectors(client);
}
