import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryOptions } from "@tanstack/svelte-query";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  type V1ListResourcesResponse,
  type RpcStatus,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data) => {
          // There may not be a prodDeployment if the project is hibernating
          return data?.prodDeployment;
        },
      },
    },
  );
}

export function useResources(
  instanceId: string,
  queryOptions?: CreateQueryOptions<
    V1ListResourcesResponse,
    ErrorType<RpcStatus>,
    V1ListResourcesResponse
  >,
) {
  const defaultOptions: CreateQueryOptions<
    V1ListResourcesResponse,
    ErrorType<RpcStatus>,
    V1ListResourcesResponse
  > = {
    select: (data) => ({
      ...data,
      // Filter out project parser and refresh triggers
      resources: data.resources.filter(
        (resource) =>
          resource.meta.name.kind !== ResourceKind.ProjectParser &&
          resource.meta.name.kind !== ResourceKind.RefreshTrigger,
      ),
    }),
    keepPreviousData: true,
  };

  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      ...defaultOptions,
      ...queryOptions,
    },
  });
}
