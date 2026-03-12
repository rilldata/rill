<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";

  export let status: V1DeploymentStatus | undefined;
  export let cancelHref: string;

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
    <LoadingSpinner />
    <h2 class="text-lg font-semibold">{getHeading(status)}</h2>
  </div>
  <Button type="secondary" onClick={handleCancel}>Cancel</Button>
</div>

<style lang="postcss">
  .loading-container {
    @apply flex flex-col items-center justify-center gap-y-6 flex-1;
  }

  .loading-content {
    @apply flex flex-col items-center gap-y-3;
  }
</style>
