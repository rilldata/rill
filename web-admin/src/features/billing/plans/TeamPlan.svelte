<script lang="ts">
  import {
    createAdminServiceCancelBillingSubscription,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import { getErrorForMutation } from "@rilldata/web-admin/client/utils";
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
  import { startPylonChat } from "@rilldata/web-common/features/help/startPylonChat";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { DateTime } from "luxon";

  export let organization: string;
  export let subscription: V1Subscription;

  $: plan = subscription.plan;

  $: planCanceller = createAdminServiceCancelBillingSubscription();
  async function handleCancelPlan() {
    await $planCanceller.mutateAsync({
      organization,
    });
    eventBus.emit("notification", {
      type: "success",
      message: "Your Team plan was canceled",
    });
    void invalidateBillingInfo(organization);
    open = false;
  }

  let open = false;

  $: error = getErrorForMutation($planCanceller);
  $: currentBillingCycleEndDate = DateTime.fromJSDate(
    new Date(subscription.currentBillingCycleEndDate),
  ).toLocaleString(DateTime.DATE_MED);
</script>

<SettingsContainer title="Team plan">
  <div slot="body">
    <div>
      Next billing cycle will start on
      <b>{getNextBillingCycleDate(subscription.currentBillingCycleEndDate)}</b>
      <PricingDetails />
      <PlanQuotas {organization} quotas={plan.quotas} />
    </div>
  </div>
  <svelte:fragment slot="contact">
    <span>For any questions,</span>
    <Button
      type="link"
      compact
      forcedStyle="padding-left:2px !important;"
      on:click={startPylonChat}
    >
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
          through <span class="font-semibold"
            >{currentBillingCycleEndDate}.</span
          >
        </AlertDialogDescription>

        {#if error}
          <div class="text-red-500 text-sm py-px">
            {error}
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
