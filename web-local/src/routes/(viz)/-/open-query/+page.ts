import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function load({ url }) {
  const rt = get(runtime);

  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ url, runtime: rt });
}
