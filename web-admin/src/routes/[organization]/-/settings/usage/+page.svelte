<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";

  $: organization = $page.params.organization;

  $: billingSub = createAdminServiceGetBillingSubscription(organization);
  $: billingUrl = $billingSub.data?.billingPortalUrl;
</script>

{#if billingUrl}
  <iframe
    src={billingUrl}
    title="Orb Billing Portal"
    class="w-full h-[600px]"
  />
{/if}
