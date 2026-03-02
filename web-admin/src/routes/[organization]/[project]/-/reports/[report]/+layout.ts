import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { connectCodeToHTTPStatus } from "@rilldata/web-common/lib/errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { getRuntimeServiceGetResourceQueryOptions } from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import { ConnectError } from "@connectrpc/connect";
import { createRuntimeClientFromLayout } from "@rilldata/web-admin/lib/runtime-client-utils";

export async function load({ params, parent }) {
  const { runtime } = await parent();
  const client = createRuntimeClientFromLayout(runtime);

  const reportData = await queryClient
    .fetchQuery(
      getRuntimeServiceGetResourceQueryOptions(client, {
        name: { kind: ResourceKind.Report, name: params.report },
      }),
    )
    .catch((e) => {
      const ce = ConnectError.from(e);
      throw error(connectCodeToHTTPStatus(ce.code), ce.rawMessage);
    });

  return {
    report: reportData.resource,
  };
}
