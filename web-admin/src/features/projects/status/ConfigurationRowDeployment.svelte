<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { deploymentChipDisplays } from "./display-utils";
  import { useProjectDeployment } from "./selectors";

  export let organization: string;
  export let project: string;

  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment, isLoading, error } = $projectDeployment);

  $: currentStatusDisplay =
    deploymentChipDisplays[
      deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
    ];
</script>

<div class="config-row">
  <div class="config-label">Deployment</div>
  <div class="config-value">
    {#if isLoading}
      <Spinner status={EntityStatus.Running} size="14px" />
    {:else if error}
      <span class="text-red-600 text-sm">Error loading deployment status</span>
    {:else}
      <div class="deployment-content">
        <div class="status-badge {currentStatusDisplay.wrapperClass}">
          <svelte:component
            this={currentStatusDisplay.icon}
            {...currentStatusDisplay.iconProps}
          />
          <span class={currentStatusDisplay.textClass}>
            {currentStatusDisplay.text}
          </span>
        </div>
        {#if deployment?.statusMessage}
          <span class="status-message">{deployment.statusMessage}</span>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .config-row {
    @apply flex items-center border-b border-slate-200;
    @apply min-h-[44px];
  }

  .config-row:last-child {
    @apply border-b-0;
  }

  .config-label {
    @apply w-[140px] flex-shrink-0 px-4 py-3;
    @apply text-sm font-medium text-gray-600;
    @apply bg-slate-50;
    @apply border-r border-slate-200;
    @apply whitespace-nowrap;
  }

  .config-value {
    @apply flex-1 px-4 py-3;
    @apply text-sm;
  }

  .deployment-content {
    @apply flex items-center gap-x-3;
  }

  .status-badge {
    @apply px-2 border rounded w-fit;
    @apply flex space-x-1 items-center;
  }

  .status-message {
    @apply text-gray-600 text-sm;
  }
</style>
