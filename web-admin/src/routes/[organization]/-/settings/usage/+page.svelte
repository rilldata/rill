<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: organization = $page.params.organization;

  $: billingSub = createAdminServiceGetBillingSubscription(organization);
  $: billingUrl = $billingSub.data?.billingPortalUrl;

  let iframeLoading = true;

  // credentialless is not standard and throws lint error, but it works on chrome and safari for now.
  const iframeProps = {
    credentialless: true,
  };
</script>

{#if $billingSub.isLoading || iframeLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{/if}
{#if billingUrl}
  <iframe
    {...iframeProps}
    src={billingUrl}
    title="Orb Billing Portal"
    class="w-full h-[1000px]"
    on:load={() => (iframeLoading = false)}
  />
{/if}
