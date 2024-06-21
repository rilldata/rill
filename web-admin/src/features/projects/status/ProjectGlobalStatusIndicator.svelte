<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { deploymentChipDisplays } from "./display-utils";
  import { useProjectDeployment } from "./selectors";

  export let organization: string;
  export let project: string;

  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment } = $projectDeployment);
  $: isDeploymentNotOk =
    deployment.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_OK;
  $: currentStatusDisplay =
    deploymentChipDisplays[
      deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
    ];

  // TODO: Detect if there are any resource or parse errors. If so, show an error icon.
</script>

{#if isDeploymentNotOk}
  <svelte:component
    this={currentStatusDisplay.icon}
    {...currentStatusDisplay.iconProps}
  />
{/if}
