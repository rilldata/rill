<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { onDestroy, onMount } from "svelte";

  export let status: V1DeploymentStatus | undefined;
  export let href: string;

  const SLOW_NOTICE_DELAY_MS = 30_000;

  let showSlowNotice = false;
  let slowNoticeTimer: ReturnType<typeof setTimeout> | undefined;

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

  onMount(() => {
    slowNoticeTimer = setTimeout(() => {
      showSlowNotice = true;
    }, SLOW_NOTICE_DELAY_MS);
  });

  onDestroy(() => {
    if (slowNoticeTimer) clearTimeout(slowNoticeTimer);
  });
</script>

<div class="loading-container">
  <div class="loading-content">
    <LoadingSpinner />
    <h2 class="text-lg font-semibold">{getHeading(status)}</h2>
    {#if showSlowNotice}
      <CtaNeedHelp leading="This is taking longer than usual." />
    {/if}
  </div>
  <Button type="secondary" {href}>Back to projects</Button>
</div>

<style lang="postcss">
  .loading-container {
    @apply flex flex-col items-center justify-center gap-y-6 flex-1;
  }

  .loading-content {
    @apply flex flex-col items-center gap-y-3;
  }
</style>
