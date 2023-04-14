<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Timer from "@rilldata/web-common/components/icons/Timer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "../../client";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: deploymentStatus = $proj.data?.productionDeployment?.status;
</script>

{#if !deploymentStatus}
  <Spacer />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
  <Timer className="text-amber-600" />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING}
  <Spinner status={EntityStatus.Running} />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_OK}
  <CheckCircle className="text-green-500" />
{:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR}
  <CancelCircle className="text-red-500" />
{/if}
