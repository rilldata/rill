<script lang="ts">
  import {
    createAdminServiceUpdateBillingSubscription,
    type V1BillingPlan,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import { getCategorisedPlans } from "@rilldata/web-admin/features/billing/plans/selectors";
  import PricingDetails from "@rilldata/web-admin/features/billing/PricingDetails.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;
  export let plan: V1BillingPlan;
  export let subscription: V1Subscription;

  $: categorisedPlans = getCategorisedPlans();
  $: trialPlan = $categorisedPlans.data.trialPlan;

  $: planUpdater = createAdminServiceUpdateBillingSubscription();
  async function handleUpgradePlan() {
    if (!trialPlan) return;

    await $planUpdater.mutateAsync({
      organization,
      data: {
        planName: trialPlan.name,
      },
    });
  }

  let open = false;
</script>

<SettingsContainer title={plan.name} titleIcon="info">
  <div slot="body">
    <div>
      Your subscription ends on {subscription.currentBillingCycleEndDate}
      <PricingDetails />
    </div>
    <PlanQuotas {organization} quotas={plan.quotas} />
  </div>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <Button type="link" compact forcedStyle="padding-left:2px !important;">
      contact us
    </Button>
  </svelte:fragment>

  <AlertDialog bind:open slot="action">
    <AlertDialogTrigger asChild let:builder>
      <Button builders={[builder]} type="primary">Cancel plan</Button>
    </AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Are you sure you want to cancel?</AlertDialogTitle>

        <AlertDialogDescription>
          If you cancel your plan, you’ll still be able to access your account
          through {subscription.currentBillingCycleEndDate}.
        </AlertDialogDescription>

        {#if $planUpdater.error}
          <div class="text-red-500 text-sm py-px">
            {$planUpdater.error.message}
          </div>
        {/if}
      </AlertDialogHeader>
      <AlertDialogFooter class="mt-3">
        <Button
          type="secondary"
          on:click={handleUpgradePlan}
          loading={$planUpdater.isLoading}
        >
          Cancel plan
        </Button>
        <Button type="primary" on:click={() => (open = false)}>
          Keep plan
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</SettingsContainer>
