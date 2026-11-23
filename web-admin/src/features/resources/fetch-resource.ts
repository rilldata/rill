import { ConnectError } from "@connectrpc/connect";
import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { connectCodeToHTTPStatus } from "@rilldata/web-common/lib/errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getRuntimeServiceGetResourceQueryOptions } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { error } from "@sveltejs/kit";

/**
 * Fetches a resource in a page load function.
 * Runtime errors are converted to SvelteKit errors with the matching HTTP status, so the error page shows them.
 */
export async function fetchResource(
  client: RuntimeClient,
  kind: ResourceKind,
  name: string,
) {
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
