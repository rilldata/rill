<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge={m.billing_plan_badge_team()} description={m.billing_team_price_desc()}>
  {#snippet info()}
    {m.billing_team_info()}
  {/snippet}

  {#snippet action()}
    <DetailedUsageLink {billingPortalUrl} />
  {/snippet}

  <p class="text-sm text-fg-tertiary mt-4 pb-4">
    {m.billing_team_details()}
  </p>
</PlanContainer>
