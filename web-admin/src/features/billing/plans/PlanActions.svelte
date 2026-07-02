<script lang="ts">
  import { PaidPlanTypes } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import CancelPlanDialog from "@rilldata/web-admin/features/billing/plans/dialog/CancelPlanDialog.svelte";
  import { SELF_SERVE_PLANS_BY_NAME } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";
  import ChoosePlanDialog from "@rilldata/web-admin/features/billing/plans/dialog/ChoosePlanDialog.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    organization,
  }: {
    organization: string;
  } = $props();

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived($subscriptionQuery?.data?.subscription?.plan?.name);
  let planType = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.planType,
  );

  let showCancel = $derived(
    PaidPlanTypes[planType] && !$categorisedIssues.data?.cancelled,
  );

  let showChangePlanDialog = $state(false);
  let showChangePlan = $derived(SELF_SERVE_PLANS_BY_NAME[planName]);

  let cancelOpen = $state(false);
</script>

<div class="plan-actions">
  {#if showCancel}
    <button class="plan-action" onclick={() => (cancelOpen = true)}>
      {m.billing_cancel_subscription()}
    </button>
  {/if}

  {#if showChangePlan}
    <button class="plan-action" onclick={() => (showChangePlanDialog = true)}>
      {m.billing_change_subscription()}
    </button>
  {/if}
</div>

<ChoosePlanDialog
  bind:open={showChangePlanDialog}
  {organization}
  type="change"
/>

<CancelPlanDialog bind:open={cancelOpen} {organization} />

<style lang="postcss">
  .plan-actions {
    @apply flex flex-row items-center justify-center gap-2;
  }

  .plan-action {
    @apply text-sm font-medium text-fg-tertiary bg-transparent border-none cursor-pointer p-0;
  }

  .plan-action:hover {
    @apply text-fg-secondary underline;
  }
</style>
