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

  let {
    organization,
    subscription,
    plan,
    billingPortalUrl,
  }: {
    organization: string;
    subscription: V1Subscription;
    plan: V1BillingPlan;
    billingPortalUrl: string | undefined;
  } = $props();

  let planCanceller = $derived(createAdminServiceCancelBillingSubscription());
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

  let open = $state(false);

  let error = $derived(getErrorForMutation($planCanceller));
  let currentBillingCycleEndDate = $derived(
    DateTime.fromJSDate(
      new Date(subscription.currentBillingCycleEndDate),
    ).toLocaleString(DateTime.DATE_MED),
  );
</script>

<SettingsContainer title={plan?.displayName}>
  <div>
    Next billing cycle will start on
    <b>{getNextBillingCycleDate(subscription.currentBillingCycleEndDate)}</b>.
    {#if billingPortalUrl}
      <div>
        <a
          href={billingPortalUrl}
          target="_blank"
          rel="noreferrer noopener"
          class="invoice-link">View Invoice</a
        >
      </div>
    {/if}
    <PlanQuotas {organization} />
  </div>
  {#snippet contact()}
    <span>For any questions,</span>
    <ContactUs />
  {/snippet}

  {#snippet action()}
    <AlertDialog bind:open>
      <AlertDialogTrigger>
        {#snippet child({ props })}
          <Button {...props} type="tertiary">Cancel plan</Button>
        {/snippet}
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Cancel your Team plan?</AlertDialogTitle>

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
          <Button type="primary" onClick={() => (open = false)}
            >Keep plan</Button
          >
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  {/snippet}
</SettingsContainer>

<style lang="postcss">
  .invoice-link {
    @apply text-sm text-primary-500 no-underline mt-2 inline-block;
  }
  .invoice-link:hover {
    @apply text-primary-600 underline;
  }
</style>
