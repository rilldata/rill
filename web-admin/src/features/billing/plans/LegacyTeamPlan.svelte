<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge="Team" description="$250/mo flat + storage">
  {#snippet info()}
    $250/mo flat rate + storage overages. 10 GB included · $25/GB over. Up to 8
    slots.
  {/snippet}

  {#snippet action()}
    <DetailedUsageLink {billingPortalUrl} />
  {/snippet}

  <p class="text-sm text-fg-tertiary mt-4 pb-4">
    Flat-rate plan. $250/mo includes up to 8 slots and 10 GB storage, with
    $25/GB for overages.
  </p>
</PlanContainer>
