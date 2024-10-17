<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatusStatus } from "@rilldata/web-admin/client/utils";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: organization = $page.params.organization;

  $: allStatus = mergedQueryStatusStatus([
    createAdminServiceGetBillingSubscription(organization),
    createAdminServiceListOrganizationBillingIssues(organization),
  ]);
</script>

{#if $allStatus.isLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{:else}
  <div class="flex flex-col w-full gap-y-5">
    <Plan {organization} />
    <Payment {organization} />
  </div>
{/if}
