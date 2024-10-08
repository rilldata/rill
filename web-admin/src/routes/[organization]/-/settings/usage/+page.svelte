<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: organization = $page.params.organization;

  $: billingSub = createAdminServiceGetBillingSubscription(organization);
  $: billingUrl = $billingSub.data?.billingPortalUrl;
</script>

{#if $billingSub.isLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{:else if billingUrl}
  <iframe
    src={billingUrl}
    title="Orb Billing Portal"
    class="w-full h-[600px]"
  />
{/if}
