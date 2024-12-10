import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../runtime-client";

export function useProjectTitle(instanceId: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: "/rill.yaml" },
    {
      query: {
        select: (data) => {
          let projectData: {
            display_name?: string;
            title?: string;
            name?: string;
          } = {};
          try {
            projectData = parse(data.blob as string, {
              logLevel: "silent",
            }) as {
              display_name?: string;
              title?: string;
              name?: string;
            };
          } catch {
            // Ignore
          }

          return String(
            projectData?.display_name ||
              projectData?.title ||
              projectData?.name ||
              "Untitled Rill Project",
          );
        },
      },
    },
  );
}
