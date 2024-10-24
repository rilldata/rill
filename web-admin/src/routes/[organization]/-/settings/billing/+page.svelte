<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ organization, showUpgrade } = data);

  $: allStatus = mergedQueryStatus([
    createAdminServiceGetBillingSubscription(organization),
    createAdminServiceListOrganizationBillingIssues(organization),
  ]);
</script>

{#if $allStatus.isLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{:else}
  <Plan {organization} {showUpgrade} />
  <Payment {organization} />
{/if}
