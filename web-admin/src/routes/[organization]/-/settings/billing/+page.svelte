<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import BillingContactSetting from "@rilldata/web-admin/features/billing/contact/BillingContactSetting.svelte";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ organization, showUpgradeDialog } = data);

  $: allStatus = mergedQueryStatus([
    createAdminServiceGetBillingSubscription(organization),
    createAdminServiceListOrganizationBillingIssues(organization),
  ]);
</script>

<!-- Both the queries are used in both Plan and Payment.
     So instead of showing 2 spinner it is better to show one at the top. -->
{#if $allStatus.isLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{:else}
  <Plan {organization} {showUpgradeDialog} />

  <!-- Usage analytics: slot usage and data overages (embed coming soon) -->
  <section class="usage-analytics">
    <h3 class="usage-title">Usage</h3>
    <div class="usage-placeholder">
      <p class="usage-text">
        Slot usage and data overage analytics will appear here.
      </p>
      <a
        href="https://docs.rilldata.com/manage/billing"
        target="_blank"
        rel="noopener noreferrer"
        class="usage-link"
      >
        Learn more about billing and usage
      </a>
    </div>
  </section>

  <Payment {organization} />
  <BillingContactSetting {organization} />
{/if}

<style lang="postcss">
  .usage-analytics {
    @apply border border-border rounded-lg p-5 mt-6;
  }
  .usage-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide mb-3;
  }
  .usage-placeholder {
    @apply flex flex-col items-center gap-3 py-8;
  }
  .usage-text {
    @apply text-sm text-fg-tertiary text-center;
  }
  .usage-link {
    @apply text-sm text-primary-500 no-underline;
  }
  .usage-link:hover {
    @apply text-primary-600 underline;
  }
</style>
