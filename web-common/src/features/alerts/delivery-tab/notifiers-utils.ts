import { createRuntimeServiceListNotifierConnectors } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export function getHasSlackConnection(client: RuntimeClient) {
  return createRuntimeServiceListNotifierConnectors(
    client,
    {},
    {
      query: {
        select: (data) => !!data.connectors?.some((c) => c.name === "slack"),
      },
    },
  );
}
