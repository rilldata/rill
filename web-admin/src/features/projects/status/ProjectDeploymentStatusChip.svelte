<script lang="ts">
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    getDashboardsForProject,
    useDashboardsStatus,
  } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { invalidateDashboardsQueries } from "@rilldata/web-admin/features/projects/invalidations";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useProjectDeploymentStatus } from "./selectors";

  export let organization: string;
  export let project: string;
  export let iconOnly = false;

  $: proj = createAdminServiceGetProject(organization, project);
  // Poll specifically for the project's deployment status
  $: projectDeploymentStatus = useProjectDeploymentStatus(
    organization,
    project,
  );
  let deploymentStatus: V1DeploymentStatus;

  $: instanceId = $proj?.data?.prodDeployment?.runtimeInstanceId;

  $: deploymentStatusFromDashboards = useDashboardsStatus(instanceId);

  const queryClient = useQueryClient();

  $: if ($projectDeploymentStatus.data) {
    const prevStatus = deploymentStatus;

    // status checking for a full invalidation should only depend on deployment status
    deploymentStatus = $projectDeploymentStatus.data;

    if (
      prevStatus &&
      prevStatus !== V1DeploymentStatus.DEPLOYMENT_STATUS_OK &&
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_OK
    ) {
      getDashboardsAndInvalidate();

      // Invalidate the queries used to compose the dashboard list in the breadcrumbs
      queryClient.invalidateQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId, {
          kind: ResourceKind.MetricsView,
        }),
      );
    }
  }

  async function getDashboardsAndInvalidate() {
    const dashboardListItems = await getDashboardsForProject($proj.data);
    const dashboardNames = dashboardListItems.map(
      (listing) => listing.meta.name.name,
    );
    return invalidateDashboardsQueries(queryClient, dashboardNames);
  }

  type StatusDisplay = {
    icon: any; // SvelteComponent
    iconProps?: {
      [key: string]: unknown;
    };
    text?: string;
    textClass?: string;
    wrapperClass?: string;
  };

  const statusDisplays: Record<V1DeploymentStatus, StatusDisplay> = {
    [V1DeploymentStatus.DEPLOYMENT_STATUS_OK]: {
      icon: CheckCircle,
      iconProps: { className: "text-primary-600 hover:text-primary-500" },
      text: "ready",
      textClass: "text-primary-600",
      wrapperClass: "bg-primary-50 border-primary-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING]: {
      icon: Spinner,
      iconProps: {
        bg: "linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)",
        className: "text-purple-600 hover:text-purple-500",
        status: EntityStatus.Running,
      },
      text: "syncing",
      textClass: "text-purple-600",
      wrapperClass: "bg-purple-50 border-purple-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR]: {
      icon: CancelCircle,
      iconProps: { className: "text-red-600 hover:text-red-500" },
      text: "error",
      textClass: "text-red-600",
      wrapperClass: "bg-red-50 border-red-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED]: {
      icon: InfoCircleFilled,
      iconProps: { className: "text-indigo-600 hover:text-indigo-500" },
      text: "not deployed",
      textClass: "text-indigo-600",
      wrapperClass: "bg-indigo-50 border-indigo-300",
    },
  };

  // Merge the status from deployment and dashboards to show the chip
  let currentStatusDisplay: StatusDisplay;
  $: if (deploymentStatus || $deploymentStatusFromDashboards?.data) {
    if (
      deploymentStatus !== V1DeploymentStatus.DEPLOYMENT_STATUS_OK ||
      !$deploymentStatusFromDashboards
    ) {
      currentStatusDisplay = statusDisplays[deploymentStatus];
    } else {
      currentStatusDisplay =
        statusDisplays[
          $deploymentStatusFromDashboards?.data ??
            V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
        ];
    }
  }
</script>

{#if $deploymentStatusFromDashboards.isFetching && !$deploymentStatusFromDashboards?.data}
  <div class="p-0.5">
    <Spinner status={EntityStatus.Running} size="16px" />
  </div>
{:else if deploymentStatus}
  {#if iconOnly}
    <svelte:component
      this={currentStatusDisplay.icon}
      {...currentStatusDisplay.iconProps}
    />
  {:else}
    <div
      class="flex space-x-1 items-center px-2 border rounded rounded-[20px] w-fit {currentStatusDisplay.wrapperClass} {iconOnly &&
        'hidden'}"
    >
      <svelte:component
        this={currentStatusDisplay.icon}
        {...currentStatusDisplay.iconProps}
      />
      <span class={currentStatusDisplay.textClass}
        >{currentStatusDisplay.text}</span
      >
    </div>
  {/if}
{:else}
  <!-- Avoid layout shift for the iconOnly instance on the homepage -->
  <Spacer />
{/if}
