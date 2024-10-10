<script lang="ts">
  import {
    createAdminServiceCancelBillingSubscription,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import { getNextBillingCycleDate } from "@rilldata/web-admin/features/billing/plans/selectors";
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
  export let subscription: V1Subscription;

  $: plan = subscription.plan;

  $: planCanceller = createAdminServiceCancelBillingSubscription();
  async function handleCancelPlan() {
    await $planCanceller.mutateAsync({
      organization,
    });
    void invalidateBillingInfo(organization);
    open = false;
  }

  let open = false;
</script>

<SettingsContainer title={plan.displayName ?? plan.name}>
  <div slot="body">
    <div>
      Next billing cycle will start on
      <b>{getNextBillingCycleDate(subscription.currentBillingCycleEndDate)}</b>
      <PricingDetails />
    </div>
    <PlanQuotas {organization} quotas={plan.quotas} />
  </div>
  <svelte:fragment slot="contact">
    <span>For any questions,</span>
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
          through <end of paid period>. </end>
        </AlertDialogDescription>

        {#if $planCanceller.error}
          <div class="text-red-500 text-sm py-px">
            {$planCanceller.error.message}
          </div>
        {/if}
      </AlertDialogHeader>
      <AlertDialogFooter class="mt-3">
        <Button
          type="secondary"
          on:click={handleCancelPlan}
          loading={$planCanceller.isLoading}
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
