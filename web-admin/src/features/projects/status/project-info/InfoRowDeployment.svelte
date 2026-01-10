<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { deploymentChipDisplays } from "../display-utils";
  import { useProjectDeployment } from "../selectors";
  import InfoRow from "./InfoRow.svelte";

  export let organization: string;
  export let project: string;

  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment, isLoading, error } = $projectDeployment);

  $: currentStatusDisplay =
    deploymentChipDisplays[
      deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
    ];
</script>

<InfoRow label="Deployment">
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
</InfoRow>

<style lang="postcss">
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
