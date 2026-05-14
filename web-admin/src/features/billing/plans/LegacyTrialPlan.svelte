<script lang="ts">
  import {
    V1BillingIssueType,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";

  let {
    organization,
    subscription,
    upgrade,
  }: {
    organization: string;
    subscription: V1Subscription;
    upgrade: () => void;
  } = $props();

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let trialIssue = $derived($categorisedIssues.data?.trial);

  let isTrialExpired = $derived(
    trialIssue?.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED ||
      trialIssue?.type ===
        V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_CREDITS_DEPLETED,
  );

  const TRIAL_DAYS = 30;
  let trialEndDate = $derived(
    trialIssue?.metadata?.onTrial?.endDate ?? subscription?.trialEndDate,
  );
  let trialDaysUsed = $derived.by(() => {
    if (!trialEndDate) return 0;
    const end = new Date(trialEndDate).getTime();
    const start = end - TRIAL_DAYS * 24 * 60 * 60 * 1000;
    const now = Date.now();
    const elapsed = Math.floor((now - start) / (24 * 60 * 60 * 1000));
    return Math.max(0, Math.min(TRIAL_DAYS, elapsed));
  });
  let trialDaysRemaining = $derived(TRIAL_DAYS - trialDaysUsed);
  let trialPercent = $derived(Math.round((trialDaysUsed / TRIAL_DAYS) * 100));
</script>

<PlanContainer
  badge="Free Trial"
  description={isTrialExpired
    ? "Trial expired · Projects hibernated"
    : "30 day free trial"}
>
  {#snippet info()}
    {#if !isTrialExpired}
      Legacy free trial · 30 days, no credit card required. Projects hibernate
      when trial ends.
    {/if}
  {/snippet}

  {#snippet action()}
    <button class="subscribe-btn" onclick={upgrade}>Upgrade to Pro</button>
  {/snippet}

  <div class="trial-section">
    <div class="flex justify-between mb-1">
      <div>
        <span class="trial-label">Days used</span>
        <p class="trial-number-used">
          {trialDaysUsed}
        </p>
      </div>
      <div class="text-right">
        <span class="trial-label">Days remaining</span>
        <p
          class="trial-number"
          class:text-green-600={trialDaysRemaining > 7}
          class:text-red-600={trialDaysRemaining <= 7}
        >
          {trialDaysRemaining}
        </p>
      </div>
    </div>
    <div class="trial-bar-bg">
      <div class="trial-bar-fill" style:width="{trialPercent}%"></div>
    </div>
    <div class="flex justify-between mt-1">
      <span class="text-xs text-fg-tertiary">
        {trialPercent}% of trial used, projects will hibernate when trial ends
      </span>
      <span class="text-xs text-fg-tertiary">30 days</span>
    </div>
    <a
      class="pricing-link"
      href="https://www.rilldata.com/pricing"
      target="_blank"
      rel="noreferrer noopener"
    >
      See pricing details
      <svg
        class="w-3 h-3"
        viewBox="0 0 12 12"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path d="M1 6h9M7.5 3l3 3-3 3" />
      </svg>
    </a>
  </div>
</PlanContainer>

<style lang="postcss">
  .trial-label {
    @apply font-sans font-semibold text-xs text-fg-tertiary;
    line-height: 100%;
  }

  .trial-number-used {
    @apply font-sans font-semibold text-2xl leading-8;
  }

  .trial-number {
    @apply font-sans font-semibold text-4xl leading-10;
  }

  .trial-section {
    @apply mt-4 pt-4 border-t flex flex-col;
  }

  .trial-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .trial-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .pricing-link {
    @apply mt-3 inline-flex items-center gap-1 self-end;
    @apply text-xs font-medium text-primary-600 no-underline;
  }

  .pricing-link:hover {
    @apply underline;
  }

  .subscribe-btn {
    @apply text-sm font-medium text-white bg-primary-500 px-5 py-2 cursor-pointer border-none rounded-none;
  }

  .subscribe-btn:hover {
    @apply bg-primary-600;
  }
</style>
