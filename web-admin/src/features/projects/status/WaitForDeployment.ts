import {
  createAdminServiceGetProject,
  V1DeploymentStatus,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/status/selectors";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  fetchResources,
  ResourceKind,
  SingletonProjectParserName,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import { derived, type Unsubscriber } from "svelte/store";

const PollTime = 1000;

export class WaitForDeployment {
  private errored = false;

  // The user is 1st redirected to `invite` page. They might take some time on there or be very quick.
  // We use this boolean to identify if a deployment already succeeded by the time user lands on dashboards page.
  private shouldRedirect = false;

  private readonly unsub: Unsubscriber;

  private static instance: WaitForDeployment | undefined;

  private constructor(
    private readonly organization: string,
    private readonly project: string,
  ) {
    this.unsub = deploymentListener(organization, project).subscribe(
      async (status) => {
        if (status.isFetching) return;
        this.unsub?.();
        if (status.error) {
          this.errored = true;
        } else if (status.data) {
          const resources = await fetchResources(queryClient, status.data);
          const visualizationErrored = resources.some(
            (r) =>
              (r.meta?.name?.kind === ResourceKind.MetricsView ||
                r.meta?.name?.kind === ResourceKind.Dashboard) &&
              !!r.meta?.reconcileError,
          );
          // only mark as errored if any visualization like metrics view or custom dashboard errored
          if (visualizationErrored) {
            this.errored = true;
          }
        }

        if (this.shouldRedirect) {
          // if the user is on dashboards page before deployment was completed.
          void this.handlePostDeployment();
        } else {
          // else do not do anything until user is out of invite page.
          this.shouldRedirect = true;
        }
      },
    );
  }

  public static create(organization: string, project: string) {
    this.instance = new WaitForDeployment(organization, project);
  }

  public static wait() {
    if (!this.instance) return;

    if (!this.instance.shouldRedirect) {
      // user landed on dashboards page before deployment is completed.
      this.instance.shouldRedirect = true;
    } else {
      // deployment is already complete.
      void this.instance.handlePostDeployment();
    }

    this.instance = undefined;
  }

  private handlePostDeployment() {
    if (this.errored) {
      eventBus.emit("notification", {
        message: "Failed to deploy project",
      });
      return;
    }

    eventBus.emit("notification", {
      message: "Project deployed",
      type: "success",
      link: {
        text: "Go to dashboards",
        href: `/${this.organization}/${this.project}`,
      },
    });
  }
}

function deploymentListener(
  organization: string,
  project: string,
): CompoundQueryResult<string> {
  return derived(
    useRefetchingProject(organization, project),
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

      const instanceId =
        projectResp.data?.prodDeployment?.runtimeInstanceId ?? "";

      derived(useRefetchingProjectParser(instanceId), (projectParserResp) => {
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
          data: instanceId,
        };
      }).subscribe(set);
    },
  );
}

function useRefetchingProject(organization: string, project: string) {
  return createAdminServiceGetProject<V1GetProjectResponse>(
    organization,
    project,
    undefined,
    {
      query: {
        ...baseGetProjectQueryOptions,
        queryClient,
      },
    },
  );
}

function useRefetchingProjectParser(instanceId: string) {
  return useResource(
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
