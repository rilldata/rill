<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: organization = $page.params.organization;

  $: billingSub = createAdminServiceGetBillingSubscription(organization);
  $: billingUrl = $billingSub.data?.billingPortalUrl;

  let iframeLoading = true;
</script>

{#if $billingSub.isLoading || iframeLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{/if}
{#if billingUrl}
  <!-- TODO: resize based on page size -->
  <iframe
    src={billingUrl}
    title="Orb Billing Portal"
    class="w-full h-[600px]"
    on:load={() => (iframeLoading = true)}
  />
{/if}
