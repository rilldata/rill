import { createRuntimeServiceListNotifierConnectors } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
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
