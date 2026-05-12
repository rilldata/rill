<script lang="ts">
  import type { V1Subscription } from "@rilldata/web-admin/client";
  import {
    getBillingCycleDates,
    getBillingStatsForOrg,
    getPlanCredits,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { formatCredit } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import CostAndUsage from "@rilldata/web-admin/features/billing/plans/modules/CostAndUsage.svelte";

  let {
    organization,
    subscription,
    billingPortalUrl,
  }: {
    organization: string;
    subscription: V1Subscription;
    billingPortalUrl: string | undefined;
  } = $props();

  let billingStats = $derived(getBillingStatsForOrg(organization));
  let dailyRunRate = $derived(
    $billingStats.prodDailyCost + $billingStats.devDailyCost,
  );

  // Pro plan credit + post-credit estimate. Available credit is hard-zero
  // until the billing usage API exposes the remaining trial credit balance.
  let planCredits = $derived(getPlanCredits(organization));
  let { availableCredit } = $derived($planCredits);
  let proEstimatedCost = $derived(Math.max(dailyRunRate - availableCredit, 0));

  // Billing cycle
  let { formattedPeriodStart, formattedPeriodEnd, formattedDueDate } = $derived(
    getBillingCycleDates(subscription),
  );
</script>

<PlanContainer badge="Pro" description="Usage based pricing">
  {#snippet action()}
    {#if billingPortalUrl}
      <a
        class="pricing-link-top"
        href={billingPortalUrl}
        target="_blank"
        rel="noreferrer noopener"
      >
        View detailed usage
        <svg
          class="w-3 h-3"
          viewBox="0 0 12 12"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M3 3h6v6M3 9l6-6" />
        </svg>
      </a>
    {/if}
  {/snippet}

  {#snippet info()}
    $0.15/unit/hr · $1/GB storage/mo. Cancel anytime.
  {/snippet}

  <div class="pro-stats">
    <div class="pro-stat-col">
      <span class="pro-stat-label">Current period estimate</span>
      <span class="pro-stat-amount text-fg-secondary"
        >{formatCredit(dailyRunRate)}</span
      >
      <span class="pro-stat-sub">
        {formattedPeriodStart} – {formattedPeriodEnd}
      </span>
    </div>
    <div class="pro-stat-col border-l pl-6">
      <span class="pro-stat-label">Available credit</span>
      <span class="pro-stat-amount text-green-700"
        >{formatCredit(availableCredit)}</span
      >
      <span class="credit-pill">
        <svg viewBox="0 0 12 12" fill="none" class="w-3 h-3 shrink-0">
          <path
            d="M10 3L4.5 8.5 2 6"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        Trial credit applied to your bill
      </span>
    </div>
    <div class="pro-stat-col border-l pl-6">
      <span class="pro-stat-label">Estimated cost after applied credit</span>
      <span class="pro-stat-amount text-fg-primary"
        >{formatCredit(proEstimatedCost)}</span
      >
      {#if formattedDueDate}
        <span class="pro-stat-sub">Due {formattedDueDate}</span>
      {/if}
    </div>
  </div>

  {#snippet footer()}
    <CostAndUsage {organization} />
  {/snippet}
</PlanContainer>

<style lang="postcss">
  .pro-stats {
    @apply flex gap-6 mt-4 pt-4 border-t;
    min-height: 92px;
  }

  .pro-stat-col {
    @apply flex flex-col gap-2 flex-1 justify-center;
  }

  .pro-stat-label {
    @apply text-xs font-semibold text-fg-tertiary;
    line-height: 1;
  }

  .pro-stat-amount {
    @apply text-4xl font-semibold leading-none;
  }

  .pro-stat-sub {
    @apply text-sm font-medium text-fg-tertiary;
    line-height: 1;
  }

  .credit-pill {
    @apply inline-flex items-center gap-1 text-sm font-medium text-green-700 bg-green-50 rounded-full px-2.5 py-1;
  }

  .pricing-link-top {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-link-top:hover {
    @apply underline;
  }
</style>
