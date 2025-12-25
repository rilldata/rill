import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import httpClient from "@rilldata/web-common/runtime-client/http-client";

export async function load({ url }) {
  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ url, instanceId: httpClient.getInstanceId() });
}
