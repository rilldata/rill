<script lang="ts" context="module">
  /**
   * 1. base  - When user chooses to upgrade from a trial plan.
   * 2. size  - When user hits the size limit and wants to upgrade.
   * 3. org   - When user hits the organization limit and wants to upgrade.
   * 4. proj  - When user hits the project limit and wants to upgrade.
   * 5. renew - After user cancels a subscription and wants to renew.
   */
  export type TeamPlanDialogTypes = "base" | "size" | "org" | "proj" | "renew";
  export type ShowTeamPlanDialogCallback = (
    type: TeamPlanDialogTypes,
    endDate: string,
  ) => void;
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import { mergedQueryStatusStatus } from "@rilldata/web-admin/client/utils";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    fetchPaymentsPortalURL,
    fetchTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { getSubscriptionResumedText } from "@rilldata/web-admin/features/billing/plans/utils";
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
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client/index.js";
  import { PopupWindow } from "@rilldata/web-common/lib/openPopupWindow";
  import { getPaymentIssues } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";

  export let organization: string;
  export let open = false;
  export let endDate = "";

  export let type: TeamPlanDialogTypes;

  let title: string;
  let description =
    "Starting a Team plan will end your trial and start your billing cycle today. " +
    "Pricing is based on amount of data ingested (and compressed) into Rill.";
  let buttonText = "Start Team plan";
  function setCopyBasedOnType(t: TeamPlanDialogTypes) {
    switch (t) {
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

      case "renew":
        title = "Renew Team plan";
        description =
          `Your billing cycle will resume ${getSubscriptionResumedText(endDate)}. ` +
          "Pricing is based on amount of data ingested (and compressed) into Rill";
        buttonText = "Continue";
        break;
    }
  }
  $: setCopyBasedOnType(type);

  $: paymentIssues = getPaymentIssues(organization);

  const userPromptWindow = new PopupWindow();

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();
  $: allStatus = mergedQueryStatusStatus([
    paymentIssues,
    planUpdater,
    planRenewer,
  ]);
  async function handleUpgradePlan() {
    // only fetch when needed to avoid hitting orb for list of plans too often
    const teamPlan = await fetchTeamPlan();
    if ($paymentIssues.data?.length) {
      await userPromptWindow.openAndWait(
        await fetchPaymentsPortalURL(
          organization,
          `${$page.url.protocol}//${$page.url.host}/-/auto-close`,
        ),
      );
    }

    if (type === "renew") {
      await $planRenewer.mutateAsync({
        organization,
        data: {
          planName: teamPlan.name,
        },
      });
    } else {
      await $planUpdater.mutateAsync({
        organization,
        data: {
          planName: teamPlan.name,
        },
      });
    }
    void invalidateBillingInfo(organization);
    open = false;
  }
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

      {#if $allStatus.isError}
        <div class="text-red-500 text-sm py-px">
          {#each $allStatus.errors as e}
            <div>{e}</div>
          {/each}
        </div>
      {/if}
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button type="secondary" on:click={() => (open = false)}>Close</Button>
      <Button
        type="primary"
        on:click={handleUpgradePlan}
        loading={$allStatus.isLoading}
      >
        {buttonText}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
