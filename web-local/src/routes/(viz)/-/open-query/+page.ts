import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";

export async function load({ url }) {
  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ url });
}
