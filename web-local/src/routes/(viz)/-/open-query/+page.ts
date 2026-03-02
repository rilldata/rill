import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import { getLocalRuntimeClient } from "../../../../lib/local-runtime-config";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ url }) {
  const query = getQueryFromUrl(url);

  await openQuery({
    query,
    client: getLocalRuntimeClient(),
  });
}
