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
  import { createAdminServiceUpdateBillingSubscription } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations.ts";
  import { getErrorForMutation } from "@rilldata/web-admin/client/utils.ts";
  import { fetchProPlan } from "@rilldata/web-admin/features/billing/plans/selectors.ts";

  let {
    open = $bindable(),
    organization,
  }: { open: boolean; organization: string } = $props();

  let proPlanUpdater = createAdminServiceUpdateBillingSubscription();
  let upgradeProLoading = $state($proPlanUpdater.isPending);
  let upgradeProError = $derived(getErrorForMutation($proPlanUpdater));

  async function confirmUpgradeToPro() {
    const teamPlan = await fetchProPlan();
    // if (!teamPlan) return;
    console.log(teamPlan);
    return;
    await $proPlanUpdater.mutateAsync({
      org: organization,
      data: { planName: teamPlan.name },
    });
    eventBus.emit("notification", {
      type: "success",
      message: "You're on the Pro plan",
    });
    void invalidateBillingInfo(organization);
    open = false;
  }
</script>

<AlertDialog
  bind:open
  onOpenChange={(newOpen) => {
    if (newOpen) {
      upgradeProError = null;
    }
  }}
>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Upgrade to Pro?</AlertDialogTitle>
      <AlertDialogDescription>
        Your subscription will start today using the payment method on file.
        You'll be billed monthly based on usage at $0.15/unit/hr and $1/GB
        storage/mo. Cancel anytime.
      </AlertDialogDescription>
      {#if upgradeProError}
        <p class="text-red-500 text-sm">{upgradeProError}</p>
      {/if}
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button
        type="secondary"
        onClick={() => (open = false)}
        disabled={upgradeProLoading}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={confirmUpgradeToPro}
        loading={upgradeProLoading}
        disabled={upgradeProLoading}
      >
        Upgrade to Pro
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
