<script lang="ts">
  import type {
    V1BillingCreditInfo,
    V1BillingPlan,
    V1Subscription,
  } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartGrowthPlanDialog from "@rilldata/web-admin/features/billing/plans/StartGrowthPlanDialog.svelte";
  import type { GrowthPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  export let subscription: V1Subscription;
  export let creditInfo: V1BillingCreditInfo | undefined;
  export let plan: V1BillingPlan;
  export let showUpgradeDialog: boolean;

  $: remaining = creditInfo?.remainingCredit ?? 0;
  $: total = creditInfo?.totalCredit ?? 250;
  $: pctUsed = total > 0 ? Math.round(((total - remaining) / total) * 100) : 0;
  $: burnRate = creditInfo?.burnRatePerDay ?? 0;
  $: daysRemaining =
    burnRate > 0 ? Math.ceil(remaining / burnRate) : undefined;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: creditExhausted = !!$categorisedIssues.data?.creditExhausted;

  let dialogOpen = showUpgradeDialog;
  $: dialogType = (
    creditExhausted ? "credit-exhausted" : pctUsed >= 80 ? "credit-low" : "base"
  ) as GrowthPlanDialogTypes;
</script>

<SettingsContainer title={plan?.displayName || "Free plan"}>
  <div slot="body">
    <div class="flex flex-col gap-3">
      <div>
        <span class="font-semibold">${remaining.toFixed(0)}</span> of
        <span>${total.toFixed(0)}</span> credit remaining ({pctUsed}% used)
      </div>

      <div class="w-full bg-gray-200 rounded-full h-2">
        <div
          class="h-2 rounded-full transition-all"
          class:bg-blue-500={pctUsed < 80}
          class:bg-yellow-500={pctUsed >= 80 && pctUsed < 95}
          class:bg-red-500={pctUsed >= 95}
          style="width: {Math.min(pctUsed, 100)}%"
        />
      </div>

      {#if daysRemaining !== undefined}
        <div class="text-sm text-gray-500">
          At current usage, credit runs out in ~{daysRemaining} days
        </div>
      {/if}

      {#if plan}
        <PlanQuotas {organization} />
      {/if}
    </div>
  </div>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <ContactUs />
  </svelte:fragment>

  <Button type="primary" slot="action" onClick={() => (dialogOpen = true)}>
    Upgrade to Growth
  </Button>
</SettingsContainer>

{#if !$categorisedIssues.isLoading}
  <StartGrowthPlanDialog bind:open={dialogOpen} {organization} type={dialogType} />
{/if}
