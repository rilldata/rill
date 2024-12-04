<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";

  export let data: PageData;

  let iframeLoading = true;

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
  src={data.billingPortalUrl}
  title="Orb Billing Portal"
  class="w-full h-[1000px]"
  on:load={() => (iframeLoading = false)}
/>
