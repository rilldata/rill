<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    formatEnvironmentName,
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import { useProjectDeployment, useRuntimeVersion } from "../selectors";

  export let organization: string;
  export let project: string;

  // Deployment data
  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment, isLoading: deploymentLoading } = $projectDeployment);
  $: deploymentStatus =
    deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
  $: deploymentEnvironment = formatEnvironmentName(deployment?.environment);

  // Version
  $: runtimeVersionQuery = useRuntimeVersion();
  $: version = $runtimeVersionQuery.data?.version || "—";
</script>

<div class="header">
  <div class="header-left">
    <h2 class="title">Project Status</h2>
    {#if deploymentLoading}
      <Spinner status={EntityStatus.Running} size="16px" />
    {:else}
      <div class="deployment-info">
        <span class="deployment-env">{deploymentEnvironment}:</span>
        <Tooltip distance={8}>
          <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">{getStatusLabel(deploymentStatus)}</p>
          </TooltipContent>
        </Tooltip>
        {#if deployment?.statusMessage}
          <span class="status-message">— {deployment.statusMessage}</span>
        {/if}
      </div>
    {/if}
  </div>
  <div class="version">
    {version}
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex items-center justify-between mb-3;
  }

  .header-left {
    @apply flex items-center gap-3;
  }

  .title {
    @apply text-lg font-semibold text-fg-primary;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full;
  }

  .version {
    @apply text-sm font-mono text-fg-secondary;
  }

  .deployment-info {
    @apply flex items-center gap-2;
  }

  .deployment-env {
    @apply text-sm font-medium text-fg-secondary italic;
  }

  .status-message {
    @apply text-sm text-fg-secondary;
    @apply max-w-md truncate;
  }

  .tooltip-text {
    @apply text-sm max-w-[200px];
  }
</style>
