import { goto } from "$app/navigation";
import {
  createAdminServiceGetProject,
  V1DeploymentStatus,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  fetchResources,
  ResourceKind,
  SingletonProjectParserName,
  useResourceV2,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import { derived, writable } from "svelte/store";

export const shouldWaitForDeployment = writable(false);

const PollTime = 1000;

export function waitForDeployment(orgName: string, projName: string) {
  void goto(`/${orgName}/${projName}`);
  const unsub = deploymentListener(orgName, projName).subscribe(
    async (status) => {
      if (status.isFetching) return;
      unsub?.();
      if (status.error) {
        void goto(`/${orgName}/${projName}/-/status`);
      } else {
        const resources = await fetchResources(queryClient, status.data);
        const metricsView = resources.find(
          (r) => r.meta?.name?.kind === ResourceKind.MetricsView,
        );
        if (metricsView?.meta?.name?.name) {
          void goto(`/${orgName}/${projName}/${metricsView.meta.name.name}`);
          return;
        }

        // if there is no metrics view, try to find a custom dashboard
        const dashboard = resources.find(
          (r) => r.meta?.name?.kind === ResourceKind.Dashboard,
        );
        if (dashboard?.meta?.name?.name) {
          void goto(
            `/${orgName}/${projName}/-/dashboards/${dashboard.meta.name.name}`,
          );
        }

        // any new visual resource would need to be added here
      }
    },
  );
}

export function deploymentListener(
  orgName: string,
  projName: string,
): CompoundQueryResult<string> {
  return derived(
    useRefetchingProject(orgName, projName),
    (projectResp, set) => {
      if (projectResp.isFetching) {
        set({
          isFetching: true,
          error: undefined,
        });
        return;
      }

      if (
        projectResp.data?.prodDeployment?.status ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR
      ) {
        set({
          isFetching: false,
          error:
            projectResp.data.prodDeployment.statusMessage ??
            "deployment failed",
        });
        return;
      }

      derived(
        useRefetchingProjectParser(
          projectResp.data?.prodDeployment?.runtimeInstanceId,
        ),
        (projectParserResp) => {
          if (
            projectParserResp.isLoading ||
            projectParserResp.data?.meta?.reconcileStatus !==
              V1ReconcileStatus.RECONCILE_STATUS_IDLE
          ) {
            return {
              isFetching: true,
              error: undefined,
            };
          }

          return {
            isFetching: false,
            error: projectParserResp.data?.meta?.reconcileError,
            data: projectResp.data?.prodDeployment?.runtimeInstanceId,
          };
        },
      ).subscribe(set);
    },
  );
}

function useRefetchingProject(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1GetProjectResponse>(
    orgName,
    projName,
    undefined,
    {
      query: {
        refetchInterval: (data) => {
          if (
            !data?.prodDeployment?.status ||
            data.prodDeployment.status ===
              V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED ||
            data.prodDeployment.status ===
              V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING
          ) {
            return PollTime;
          }
          return false;
        },
      },
    },
  );
}

function useRefetchingProjectParser(instanceId: string) {
  return useResourceV2(
    instanceId,
    SingletonProjectParserName,
    ResourceKind.ProjectParser,
    {
      queryClient,
      refetchInterval: (data) => {
        if (
          !data?.meta?.reconcileStatus ||
          data.meta.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_RUNNING ||
          data.meta.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_PENDING
        ) {
          return PollTime;
        }
        return false;
      },
    },
  );
}
