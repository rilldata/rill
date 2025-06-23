import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";
import { createLocalServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service.ts";
import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../runtime-client";
import { derived } from "svelte/store";

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

export function getRequestProjectAccessUrl(project: Project) {
  const url = new URL(project.frontendUrl);
  url.pathname = "/-/request-project-access";
  url.searchParams.set("organization", project.orgName);
  url.searchParams.set("project", project.name);
  // Since this already has a user action, skip showing a "request" button in cloud as well.
  // Adding `auto_request` is handled there.
  url.searchParams.set("auto_request", "true");
  return url.toString();
}
