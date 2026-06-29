<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { formatCredit } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { getPlanCredits } from "@rilldata/web-admin/features/billing/plans/selectors.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
  import { PricingDetailsCompact } from "@rilldata/web-common/features/billing/pricing-details.ts";
  import PricingLink from "@rilldata/web-admin/features/billing/plans/modules/PricingLink.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    organization,
    upgrade,
  }: {
    organization: string;
    upgrade: () => void;
  } = $props();

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let trialIssue = $derived($categorisedIssues.data?.trial);

  let planCredits = $derived(getPlanCredits(organization, trialIssue));
  let { usedCredit, availableCredit, creditPercent } = $derived($planCredits);
</script>

<PlanContainer badge={m.billing_plan_badge_pro_trial()} description={m.billing_free_credit_desc()}>
  {#snippet info()}
    {m.billing_no_time_limit()}<br />
    {PricingDetailsCompact}<br />
    {m.billing_unit_spec()}
  {/snippet}

  {#snippet action()}
    <button class="subscribe-btn" onclick={upgrade}>{m.billing_upgrade_to_pro()}</button>
  {/snippet}

  <div class="credit-section">
    <div class="flex justify-between">
      <span class="credit-label">{m.billing_used_credit()}</span>
      <span class="credit-label">{m.billing_available_credit()}</span>
    </div>
    <div class="flex justify-between items-end">
      <span class="credit-used">{formatCredit(usedCredit)}</span>
      <span class="credit-available">{formatCredit(availableCredit)}</span>
    </div>
    <div class="credit-bar-bg">
      <div class="credit-bar-fill" style:width="{creditPercent}%"></div>
    </div>
    <span class="text-xs text-fg-tertiary font-medium">
      {m.billing_credit_percent_used({ percent: String(creditPercent) })}
    </span>

    <PricingLink />
  </div>
</PlanContainer>

<style lang="postcss">
  .credit-section {
    @apply flex flex-col gap-2 mt-4 pt-4 border-t;
  }

  .credit-label {
    @apply text-xs font-semibold text-fg-tertiary;
    line-height: 1;
  }

  .credit-used {
    @apply text-2xl font-semibold text-fg-tertiary;
  }

  .credit-available {
    @apply text-4xl font-semibold text-green-600;
  }

  .credit-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .credit-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .subscribe-btn {
    @apply text-sm font-medium text-white bg-primary-500 px-5 py-2 cursor-pointer border-none rounded-none;
  }

  .subscribe-btn:hover {
    @apply bg-primary-600;
  }
</style>
