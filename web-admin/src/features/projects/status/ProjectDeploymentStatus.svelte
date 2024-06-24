<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { deploymentChipDisplays } from "./display-utils";
  import { useProjectDeployment } from "./selectors";

  export let organization: string;
  export let project: string;

  // Poll specifically for the project's deployment
  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment, isLoading, error } = $projectDeployment);

  $: currentStatusDisplay =
    deploymentChipDisplays[
      deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
    ];
</script>

<section class="deployment-status">
  <h3 class="deployment-label">Deployment</h3>
  {#if isLoading}
    <div class="py-1">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:else if error}
    <div class="py-0.5">
      <span class="text-red-600">Error loading deployment status</span>
    </div>
  {:else}
    <div
      class="deployment-status-tag-wrapper {currentStatusDisplay.wrapperClass}"
    >
      <svelte:component
        this={currentStatusDisplay.icon}
        {...currentStatusDisplay.iconProps}
      />
      <span class={currentStatusDisplay.textClass}>
        {currentStatusDisplay.text}
      </span>
    </div>
    {#if deployment?.statusMessage}
      {deployment.statusMessage}
    {/if}
  {/if}
</section>

<style lang="postcss">
  .deployment-status {
    @apply flex flex-col gap-y-1;
  }

  .deployment-label {
    @apply text-[10px] leading-none font-semibold uppercase;
    @apply text-gray-500;
  }

  .deployment-status-tag-wrapper {
    @apply px-2 border rounded w-fit;
    @apply flex space-x-1 items-center;
  }
</style>
