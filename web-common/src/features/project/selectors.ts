import { parseDocument } from "yaml";
import { createRuntimeServiceGetFile } from "../../runtime-client";

export function useProjectTitle(instanceId: string) {
  return createRuntimeServiceGetFile(instanceId, "rill.yaml", {
    query: {
      select: (data) => {
        const projectData = parseDocument(data.blob)?.toJS();
        return projectData?.title ?? projectData?.name;
      },
    },
  });
}
