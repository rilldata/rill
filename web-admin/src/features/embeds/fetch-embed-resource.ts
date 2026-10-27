import { ConnectError } from "@connectrpc/connect";
import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { connectCodeToHTTPStatus } from "@rilldata/web-common/lib/errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getRuntimeServiceGetResourceQueryOptions } from "@rilldata/web-common/runtime-client";
import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2/context";
import { error } from "@sveltejs/kit";

/**
 * Fetches a resource in a page load function under `/-/embed`.
 * Embed routes don't have the project layout, so the runtime client is built from the embed layout's data.
 */
export async function fetchEmbedResource(
  {
    instanceId,
    runtimeHost,
    accessToken,
  }: { instanceId: string; runtimeHost: string; accessToken: string },
  kind: ResourceKind,
  name: string,
) {
  // The embed layout shows an error for missing params, so there's nothing to fetch.
  if (!instanceId || !runtimeHost || !accessToken) throw error(400);

  const client = getRuntimeClient({
    host: runtimeHost,
    instanceId,
    jwt: accessToken,
    authContext: "embed",
  });

  const resp = await queryClient
    .fetchQuery(
      getRuntimeServiceGetResourceQueryOptions(client, {
        name: { kind, name },
      }),
    )
    .catch((e) => {
      const ce = ConnectError.from(e);
      throw error(connectCodeToHTTPStatus(ce.code), ce.rawMessage);
    });
  return resp.resource;
}
