<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
  import { invalidateProjectQueries } from "@rilldata/web-admin/components/projects/invalidations";
  import { useProject } from "@rilldata/web-admin/components/projects/use-project";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Timer from "@rilldata/web-common/components/icons/Timer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let organization: string;
  export let project: string;

  const queryClient = useQueryClient();

  $: proj = useProject(organization, project);
  let deploymentStatus: V1DeploymentStatus;

  $: if ($proj.data?.prodDeployment?.status) {
    const prevStatus = deploymentStatus;

    deploymentStatus = $proj.data?.prodDeployment?.status;

    if (
      prevStatus !== V1DeploymentStatus.DEPLOYMENT_STATUS_OK &&
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_OK
    ) {
      getDashboardsAndInvalidate();
    }
  }

  async function getDashboardsAndInvalidate() {
    return invalidateProjectQueries(
      queryClient,
      await getDashboardsForProject($proj)
    );
  }
</script>

{#if !deploymentStatus}
  <Spacer />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
  <Timer className="text-amber-600 hover:text-amber-500" />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING}
  <Spinner status={EntityStatus.Running} />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_OK}
  <CheckCircle className="text-blue-500 hover:text-blue-400" />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR}
  <CancelCircle className="text-red-500 hover:text-red-400" />
{/if}
