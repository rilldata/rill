import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";
import {
  createLocalServiceGetProjectRequest,
  createLocalServiceGitStatus,
  getLocalServiceGithubRepoStatusQueryOptions,
} from "@rilldata/web-common/runtime-client/local-service.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
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

export function getManageProjectAccess(orgName: string, projectName: string) {
  return derived(
    createLocalServiceGetProjectRequest(orgName ?? "", projectName ?? "", {
      query: {
        enabled: !!projectName,
      },
    }),
    (selectedProjectResp) =>
      Boolean(selectedProjectResp.data?.projectPermissions?.manageProject),
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
  url.searchParams.set("role", ProjectUserRoles.Admin);
  return url.toString();
}

export function getLocalGitRepoStatus() {
  const gitRepoOptions = derived(createLocalServiceGitStatus(), (gitStatus) =>
    getLocalServiceGithubRepoStatusQueryOptions(
      gitStatus.data?.githubUrl ?? "",
      {
        query: {
          enabled: !!gitStatus.data?.githubUrl && !gitStatus.data?.managedGit,
        },
      },
    ),
  );

  return createQuery(gitRepoOptions);
}
