<script lang="ts">
  import { getCategorisedPlans } from "@rilldata/web-admin/features/billing/plans/selectors";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import PricingDetails from "@rilldata/web-admin/features/billing/PricingDetails.svelte";
  import { createAdminServiceUpdateBillingSubscription } from "@rilldata/web-admin/client/index.js";

  export let organization: string;
  export let open = false;

  const categorisedPlans = getCategorisedPlans();
  $: teamPlan = $categorisedPlans.data.teamPlan;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  async function handleUpgradePlan() {
    if (!teamPlan) return;

    await $planUpdater.mutateAsync({
      organization,
      data: {
        planName: teamPlan.name,
      },
    });
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Start Team plan</AlertDialogTitle>

      <AlertDialogDescription>
        Your trial will end and your billing cycle will start today. Pricing is
        based on amount of data ingested (and compressed) into Rill.
        <PricingDetails />
        <ul>
          <li>Starts at $250/month with 10 GB included, $25/GB thereafter</li>
          <li>Unlimited projects, limited to 50 GB each</li>
        </ul>
      </AlertDialogDescription>

      {#if $planUpdater.error}
        <div class="text-red-500 text-sm py-px">
          {$planUpdater.error.message}
        </div>
      {/if}
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button type="secondary" on:click={() => (open = false)}>Close</Button>
      <Button
        type="primary"
        on:click={handleUpgradePlan}
        loading={$planUpdater.isLoading}
      >
        Continue
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
