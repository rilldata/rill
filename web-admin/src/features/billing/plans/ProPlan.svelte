<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import {
    PricingDetails,
    PricingDetailsCompact,
  } from "@rilldata/web-common/features/billing/pricing-details.ts";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge="Pro" description="Usage based pricing">
  {#snippet action()}
    <DetailedUsageLink {billingPortalUrl} />
  {/snippet}

  {#snippet info()}
    {PricingDetailsCompact} Cancel anytime.
  {/snippet}

  <div class="text-sm text-fg-tertiary mt-4 pb-4">
    {PricingDetails}
  </div>
</PlanContainer>
