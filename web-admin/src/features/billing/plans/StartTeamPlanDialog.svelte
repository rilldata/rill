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
  import {
    createAdminServiceUpdateBillingSubscription,
    type RpcStatus,
  } from "@rilldata/web-admin/client/index.js";
  import type { AxiosError } from "axios";

  export let organization: string;
  export let open = false;
  /**
   * 1. base - When user chooses to upgrade from a trial plan.
   * 2. size - When user hits the size limit and wants to upgrade.
   * 3. org - When user hits the organization limit and wants to upgrade.
   * 4. proj - When user hits the project limit and wants to upgrade.
   */
  export let type: "base" | "size" | "org" | "proj";

  let title: string;
  let description =
    "Starting a Team plan will end your trial and start your billing cycle today. " +
    "Pricing is based on amount of data ingested (and compressed) into Rill.";
  let buttonText = "Start Team plan";

  $: {
    switch (type) {
      case "base":
        title = "Start Team plan";
        buttonText = "Continue";
        break;

      case "size":
        title = "Deploying more than 10GB requires a Team plan";
        break;

      case "org":
        title = "To create another organization, start a Team plan";
        description =
          "Pricing is based on amount of data ingested (and compressed) into Rill.";
        break;

      case "proj":
        title = "To deploy a second project, start a Team plan";
        break;
    }
  }

  const categorisedPlans = getCategorisedPlans();
  $: teamPlan = $categorisedPlans.data?.teamPlan;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  async function handleUpgradePlan() {
    if (!teamPlan) return;

    await $planUpdater.mutateAsync({
      organization,
      data: {
        planName: teamPlan.name,
      },
    });
    open = false;
  }

  $: error =
    ($planUpdater.error as unknown as AxiosError<RpcStatus>)?.response?.data
      ?.message ?? $planUpdater.error?.message;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>

      <AlertDialogDescription>
        {description}
        <PricingDetails />
        <ul class="mt-5 ml-5 list-disc">
          <li>Starts at $250/month with 10 GB included, $25/GB thereafter</li>
          <li>Unlimited projects, limited to 50 GB each</li>
        </ul>
      </AlertDialogDescription>

      {#if error}
        <div class="text-red-500 text-sm py-px">
          {error}
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
        {buttonText}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
