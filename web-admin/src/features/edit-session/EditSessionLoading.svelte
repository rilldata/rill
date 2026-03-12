<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { AlertCircleIcon } from "lucide-svelte";

  export let status: V1DeploymentStatus | undefined;
  export let statusMessage: string | undefined;
  export let cancelHref: string;

  $: hasError = statusMessage?.startsWith("Provisioning failed:");
  $: heading = hasError
    ? "Failed to provision editing environment"
    : getHeading(status);
  $: errorDetail = hasError
    ? statusMessage!.replace("Provisioning failed: ", "")
    : null;

  function getHeading(s: V1DeploymentStatus | undefined): string {
    switch (s) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
        return "Provisioning your editing environment...";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
        return "Updating your editing environment...";
      default:
        return "Starting your editing environment...";
    }
  }

  function handleCancel() {
    // Full page navigation bypasses the branch injection hook in the
    // parent layout, which would otherwise re-inject @branch into the URL.
    window.location.href = cancelHref;
  }
</script>

<div class="loading-container">
  <div class="loading-content">
    {#if hasError}
      <AlertCircleIcon size="32" class="text-red-500" />
    {:else}
      <LoadingSpinner />
    {/if}
    <h2 class="text-lg font-semibold">{heading}</h2>
    {#if errorDetail}
      <p class="text-sm text-fg-muted max-w-md text-center">{errorDetail}</p>
    {/if}
  </div>
  <Button type="secondary" onClick={handleCancel}>
    {hasError ? "Go back" : "Cancel"}
  </Button>
</div>

<style lang="postcss">
  .loading-container {
    @apply flex flex-col items-center justify-center gap-y-6 flex-1;
  }

  .loading-content {
    @apply flex flex-col items-center gap-y-3;
  }
</style>
