<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { formatCredit } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { getPlanCredits } from "@rilldata/web-admin/features/billing/plans/selectors.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";

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

<PlanContainer badge="Pro Trial" description="$250 free credit">
  {#snippet info()}
    No time limit, use it until it's gone.<br />
    $0.15/unit/hr · $1/GB storage/mo<br />
    1 unit = 4GiB RAM, 1vGPU
  {/snippet}

  {#snippet action()}
    <button class="subscribe-btn" onclick={upgrade}>Upgrade to Pro</button>
  {/snippet}

  <div class="credit-section">
    <div class="flex justify-between">
      <span class="credit-label">Used credit</span>
      <span class="credit-label">Available credit</span>
    </div>
    <div class="flex justify-between items-end">
      <span class="credit-used">{formatCredit(usedCredit)}</span>
      <span class="credit-available">{formatCredit(availableCredit)}</span>
    </div>
    <div class="credit-bar-bg">
      <div class="credit-bar-fill" style:width="{creditPercent}%"></div>
    </div>
    <span class="text-xs text-fg-tertiary font-medium">
      {creditPercent}% used, projects will hibernate when credits run out.
    </span>
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

  .pricing-link {
    @apply mt-3 inline-flex items-center gap-1 self-end;
    @apply text-xs font-medium text-primary-600 no-underline;
  }

  .pricing-link:hover {
    @apply underline;
  }
</style>
