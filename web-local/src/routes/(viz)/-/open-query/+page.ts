import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import {
  LOCAL_HOST,
  LOCAL_INSTANCE_ID,
} from "../../../../lib/local-runtime-config";

export async function load({ url }) {
  await openQuery({
    url,
    runtime: { host: LOCAL_HOST, instanceId: LOCAL_INSTANCE_ID },
  });
}
