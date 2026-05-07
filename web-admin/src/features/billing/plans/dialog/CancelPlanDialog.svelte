<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.ts";
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import { getPlanTierForSubscription } from "@rilldata/web-admin/features/billing/plans/selectors.ts";
  import {
    createAdminServiceCancelBillingSubscription,
    createAdminServiceGetBillingSubscription,
    V1BillingIssueType,
  } from "@rilldata/web-admin/client";
  import { getErrorForMutation } from "@rilldata/web-admin/client/utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations.ts";

  let {
    open = $bindable(false),
    organization,
  }: {
    open: boolean;
    organization: string;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);
  let currentPlan = $derived(getPlanTierForSubscription(subscription));
  let cycleEnd = $derived(subscription?.currentBillingCycleEndDate);

  // Cancel subscription
  let planCanceller = $derived(createAdminServiceCancelBillingSubscription());
  let cancelError = $derived(getErrorForMutation($planCanceller));
  let cycleEndFormatted = $derived.by(() => {
    if (!cycleEnd) return "";
    return new Date(cycleEnd).toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      year: "numeric",
    });
  });
  async function handleCancelPlan() {
    await $planCanceller.mutateAsync({ org: organization });
    eventBus.emit("notification", {
      type: "success",
      message: `Your ${currentPlan === "pro" ? "Pro" : "Team"} plan was cancelled`,
    });
    void invalidateBillingInfo(organization, [
      V1BillingIssueType.BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED,
    ]);
    open = false;
  }
</script>

<AlertDialog
  bind:open
  onOpenChange={(newOpen) => {
    if (newOpen) {
      cancelError = null;
    }
  }}
>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        Cancel your {currentPlan === "pro" ? "Pro" : "Team"} plan?
      </AlertDialogTitle>
      <AlertDialogDescription>
        If you cancel your plan, you'll still be able to access your account
        through
        <span class="font-semibold">{cycleEndFormatted}.</span>
      </AlertDialogDescription>
      {#if cancelError}
        <p class="text-red-500 text-sm">{cancelError}</p>
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
