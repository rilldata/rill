import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { fetchProjectDeploymentDetails } from "@rilldata/web-admin/features/projects/selectors.ts";
import type { CreateBaseMutationResult } from "@tanstack/svelte-query";
import type {
  AdminServiceCreateProjectBody,
  RpcStatus,
  V1CreateProjectResponse,
} from "@rilldata/web-admin/client";
import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client.ts";
import { createRuntimeServiceUnpackEmptyMutation } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

export async function createProjectAndInitEmpty(
  createProjectMutation: CreateBaseMutationResult<
    V1CreateProjectResponse,
    RpcStatus,
    {
      org: string;
      data: AdminServiceCreateProjectBody;
    },
    unknown
  >,
  organization: string,
  project: string,
  displayName: string,
) {
  const resp = await createProjectMutation.mutateAsync({
    org: organization,
    data: {
      project,
      generateManagedGit: true,
      prodSlots: "4",
    },
  });
  // No need to wait for the project to be created if displayName is the same as project name
  if (project === displayName) return resp.project?.frontendUrl;

  // Wait for deployment to be created and instace to be ready.
  let deployment:
    | Awaited<ReturnType<typeof fetchProjectDeploymentDetails>>
    | undefined = undefined;
  await asyncWaitUntil(
    async () => {
      deployment = await fetchProjectDeploymentDetails(
        organization,
        project,
        undefined,
      );
      return !!deployment.runtime.host;
    },
    1000 * 60 * 5,
    1000,
  );
  if (!deployment)
    throw new Error(
      `Project ${project} in organization ${organization} failed to initialize`,
    );

  const runtimeClient = getCloudRuntimeClient(deployment.runtime);
  const unpackEmptyMutation = createRuntimeServiceUnpackEmptyMutation(
    runtimeClient,
    undefined,
    queryClient,
  );
  // Unpack empty proejct with the supplied display name
  await get(unpackEmptyMutation).mutateAsync({
    displayName,
    olap: "duckdb",
  });

  return resp.project?.frontendUrl;
}
