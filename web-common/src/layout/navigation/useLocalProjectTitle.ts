import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { parseDocument } from "yaml";

export function useLocalProjectTitle(instanceId: string) {
  return createRuntimeServiceGetFile(instanceId, "rill.yaml", {
    query: {
      select: (data) => {
        if (!data?.blob) return "";

        const json = parseDocument(data.blob).toJS();
        return json.title ?? json.name;
      },
    },
  });
}
