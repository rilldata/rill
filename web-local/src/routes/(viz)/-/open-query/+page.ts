import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import {
  LOCAL_HOST,
  LOCAL_INSTANCE_ID,
} from "../../../../lib/local-runtime-config";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ url }) {
  const query = getQueryFromUrl(url);

  await openQuery({
    query,
    runtime: { host: LOCAL_HOST, instanceId: LOCAL_INSTANCE_ID },
  });
}
