import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryOptions } from "@tanstack/svelte-query";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1ListResourcesResponse,
  RpcStatus,
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

type ResourcesResponse = V1ListResourcesResponse & { timestamp: number };

export function useResources(
  instanceId: string,
  queryOptions?: CreateQueryOptions<
    V1ListResourcesResponse,
    ErrorType<RpcStatus>,
    ResourcesResponse
  >,
) {
  const defaultOptions: CreateQueryOptions<
    V1ListResourcesResponse,
    ErrorType<RpcStatus>,
    ResourcesResponse
  > = {
    select: (data) => ({
      ...data,
      resources: data.resources.filter(
        (resource) =>
          resource.meta.name.kind !== ResourceKind.ProjectParser &&
          resource.meta.name.kind !== ResourceKind.RefreshTrigger,
      ),
      timestamp: Date.now(),
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
