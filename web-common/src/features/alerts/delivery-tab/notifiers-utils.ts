import { createRuntimeServiceListNotifierConnectors } from "@rilldata/web-common/runtime-client";

export function getHasSlackConnection(runtimeId: string) {
  return createRuntimeServiceListNotifierConnectors(runtimeId, {
    query: {
      select: (data) => {
        console.log(data);
        return !!data.connectors?.some((c) => c.name === "slack");
      },
    },
  });
}
