<script lang="ts">
  import {
    createAdminServiceCancelBillingSubscription,
    V1BillingIssueType,
    type V1BillingPlan,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import { getErrorForMutation } from "@rilldata/web-admin/client/utils";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import { getNextBillingCycleDate } from "@rilldata/web-admin/features/billing/plans/selectors";
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { DateTime } from "luxon";

  export let organization: string;
  export let subscription: V1Subscription;
  export let plan: V1BillingPlan;

  $: planCanceller = createAdminServiceCancelBillingSubscription();
  async function handleCancelPlan() {
    await $planCanceller.mutateAsync({
      org: organization,
    });
    eventBus.emit("notification", {
      type: "success",
      message: "Your Team plan was cancelled",
    });
    void invalidateBillingInfo(organization, [
      V1BillingIssueType.BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED,
    ]);
    open = false;
  }

  let open = false;

  $: error = getErrorForMutation($planCanceller);
  $: currentBillingCycleEndDate = DateTime.fromJSDate(
    new Date(subscription.currentBillingCycleEndDate),
  ).toLocaleString(DateTime.DATE_MED);
</script>

<SettingsContainer title={plan?.displayName}>
  <div slot="body">
    Next billing cycle will start on
    <b>{getNextBillingCycleDate(subscription.currentBillingCycleEndDate)}</b>.
    <a
      href="https://www.rilldata.com/pricing"
      target="_blank"
      rel="noreferrer noopener">See pricing details -></a
    >
    <PlanQuotas {organization} />
  </div>
  <svelte:fragment slot="contact">
    <span>For any questions,</span>
    <ContactUs />
  </svelte:fragment>

  <AlertDialog bind:open slot="action">
    <AlertDialogTrigger asChild let:builder>
      <Button builders={[builder]} type="secondary" gray>Cancel plan</Button>
    </AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Are you sure you want to cancel?</AlertDialogTitle>

        <AlertDialogDescription>
          If you cancel your plan, you'll still be able to access your account
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
          onClick={handleCancelPlan}
          loading={$planCanceller.isPending}
        >
          Cancel plan
        </Button>
        <Button type="primary" onClick={() => (open = false)}>Keep plan</Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</SettingsContainer>
