import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client.ts";

export async function load({ url: { searchParams }, parent }) {
  const { runtime } = await parent();
  const client = getCloudRuntimeClient(runtime);

  // Fetch connectors and wait for the data to be loaded into cache
  await fetchAnalyzeConnectors(client);

  return {
    schema: searchParams.get("schema") ?? undefined,
  };
}
