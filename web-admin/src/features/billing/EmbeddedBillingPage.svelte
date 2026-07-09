<script lang="ts">
  import Spinner from "web-common/src/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "web-common/src/features/entity-management/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
  title={m.billing_orb_portal_title()}
  class="w-full h-[1000px]"
  onload={() => (iframeLoading = false)}
></iframe>

<style lang="postcss">
</style>
