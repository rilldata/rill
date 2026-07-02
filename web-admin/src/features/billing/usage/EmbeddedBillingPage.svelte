<script lang="ts">
  import Spinner from "web-common/src/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "web-common/src/features/entity-management/types";

  let { billingPortalUrl }: { billingPortalUrl: string } = $props();

  let iframeLoading = $state(true);

  // credentialless is not standard and throws lint error, but it works on chrome and safari for now.
  const iframeProps = {
    credentialless: true,
  };
</script>

{#if iframeLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{/if}

<iframe
  {...iframeProps}
  src={billingPortalUrl}
  title="Orb Billing Portal"
  class="w-full h-[1000px]"
  onload={() => (iframeLoading = false)}
></iframe>

<style lang="postcss">
</style>
