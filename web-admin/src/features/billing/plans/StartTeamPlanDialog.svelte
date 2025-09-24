<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
  } from "@rilldata/web-admin/client/index.js";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    fetchPaymentsPortalURL,
    fetchTeamPlan,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import {
    getSubscriptionResumedText,
    showWelcomeToRillDialog,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
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
  import PricingDetails from "@rilldata/web-common/features/billing/PricingDetails.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let open = false;
  export let endDate = "";

  export let type: TeamPlanDialogTypes;

  let title: string;
  let description =
    "Starting a Team plan will end your trial and start your billing cycle today.";
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
        description = "";
        break;

      case "proj":
        title = "To deploy a second project, start a Team plan";
        break;

      case "renew":
        title = "Renew Team plan";
        description = `Your billing cycle will resume ${getSubscriptionResumedText(endDate)}.`;
        buttonText = "Continue";
        break;

      case "trial-expired":
        title = "Start Team plan";
        description =
          "Starting Team plan will wake your projects and start your billing cycle today.";
        buttonText = "Continue";
        break;
    }
  }
  $: setCopyBasedOnType(type);

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;
  $: redirect = $page.url.searchParams.get("redirect");

  let loading = false;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();
  $: allStatus = mergedQueryStatus([
    categorisedIssues,
    planUpdater,
    planRenewer,
  ]);
  async function handleUpgradePlan() {
    loading = true;
    // only fetch when needed to avoid hitting orb for list of plans too often
    const teamPlan = await fetchTeamPlan();
    if (paymentIssues?.length) {
      window.open(
        await fetchPaymentsPortalURL(
          organization,
          getBillingUpgradeUrl($page, organization),
        ),
        "_self",
      );
      return;
    }
    loading = false;

    if (type === "renew") {
      await $planRenewer.mutateAsync({
        org: organization,
        data: {
          planName: teamPlan.name,
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Your Team plan was renewed",
      });
    } else {
      await $planUpdater.mutateAsync({
        org: organization,
        data: {
          planName: teamPlan.name,
        },
      });
      showWelcomeToRillDialog.set(true);
    }
    void invalidateBillingInfo(organization);
    open = false;
    if (redirect) {
      // redirect param could be on a different domain like the rill developer instance
      // so using goto won't work
      window.open(redirect, "_self");
    }
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
        <PricingDetails extraText={description} />
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
      <Button type="secondary" onClick={() => (open = false)}>Close</Button>
      <Button
        type="primary"
        onClick={handleUpgradePlan}
        loading={loading || $allStatus.isLoading}
      >
        {buttonText}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
