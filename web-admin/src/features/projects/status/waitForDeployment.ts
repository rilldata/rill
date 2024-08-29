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
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { derived, type Unsubscriber } from "svelte/store";

const PollTime = 1000;

export class WaitForDeployment {
  private errored = false;
  private redirectToResource: V1Resource | undefined;

  // The user is 1st redirected to `invite` page. They might take some time on there or be very quick.
  // We use this boolean to identify if a deployment already succeeded by the time user lands on dashboards page.
  private shouldRedirect = false;

  private readonly unsub: Unsubscriber;

  private static instance: WaitForDeployment;

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
        } else {
          const resources = await fetchResources(queryClient, status.data);
          // prefer a metrics view, if there are none select a custom dashboard
          this.redirectToResource =
            resources.find(
              (r) => r.meta?.name?.kind === ResourceKind.MetricsView,
            ) ??
            resources.find(
              (r) => r.meta?.name?.kind === ResourceKind.Dashboard,
            );
        }

        if (this.shouldRedirect) {
          // if the user is on dashboards page before deployment was completed.
          this.handlePostDeployment();
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
      this.instance.handlePostDeployment();
    }

    this.instance = undefined;
  }

  private handlePostDeployment() {
    if (this.errored) {
      eventBus.emit("notification", {
        message: "Failed to deploy project",
      });
      void goto(`/${this.organization}/${this.project}`);
    } else if (this.redirectToResource) {
      let dashboardLink = "";
      switch (this.redirectToResource.meta?.name?.kind) {
        case ResourceKind.MetricsView:
          dashboardLink = `/${this.organization}/${this.project}/${this.redirectToResource.meta.name.name}`;
          break;

        case ResourceKind.Dashboard:
          dashboardLink = `/${this.organization}/${this.project}/-/dashboards/${this.redirectToResource.meta.name.name}`;
          break;

        // any new visual resource would need to be added here
      }

      eventBus.emit("notification", {
        message: "Project shouldRedirect successfully",
        ...(dashboardLink
          ? {
              link: {
                text: "Go to dashboard",
                href: dashboardLink,
              },
            }
          : {}),
      });
    }
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

function useRefetchingProject(organization: string, project: string) {
  return createAdminServiceGetProject<V1GetProjectResponse>(
    organization,
    project,
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
        queryClient,
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
