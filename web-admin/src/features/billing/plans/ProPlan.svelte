<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import {
    PricingDetails,
    PricingDetailsCompact,
  } from "@rilldata/web-common/features/billing/pricing-details.ts";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge={m.billing_plan_badge_pro()} description={m.billing_usage_based_pricing()}>
  {#snippet action()}
    <DetailedUsageLink {billingPortalUrl} />
  {/snippet}

  {#snippet info()}
    {PricingDetailsCompact} {m.billing_cancel_anytime()}
  {/snippet}

  <div class="text-sm text-fg-tertiary mt-4 pb-4">
    {PricingDetails}
  </div>
</PlanContainer>
