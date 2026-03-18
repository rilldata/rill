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
    fetchGrowthPlan,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type { GrowthPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let open = false;
  export let endDate = "";
  export let type: GrowthPlanDialogTypes;

  let title: string;
  let description: string;
  let buttonText = "Upgrade to Growth";
  function setCopyBasedOnType(t: GrowthPlanDialogTypes) {
    switch (t) {
      case "base":
        title = "Upgrade to Growth";
        description =
          "Growth is pure usage-based billing with no base fee. You pay only for the slots and storage you use.";
        buttonText = "Continue";
        break;

      case "credit-low":
        title = "Your free credit is running low";
        description =
          "Upgrade to Growth to keep your projects running after your credit is exhausted.";
        buttonText = "Upgrade to Growth";
        break;

      case "credit-exhausted":
        title = "Your free credit is exhausted";
        description =
          "Your projects have been hibernated. Upgrade to Growth to wake them and resume access.";
        buttonText = "Upgrade to Growth";
        break;

      case "renew":
        title = "Renew Growth plan";
        description = `Your billing cycle will resume ${getSubscriptionResumedText(endDate)}.`;
        buttonText = "Continue";
        break;
    }
  }
  $: setCopyBasedOnType(type);

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: paymentIssues = $categorisedIssues.data?.payment;
  $: redirect = $page.url.searchParams.get("redirect");

  let loading = false;
  let fetchError: string | null = null;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();
  $: allStatus = mergedQueryStatus([
    categorisedIssues,
    planUpdater,
    planRenewer,
  ]);

  async function handleUpgradePlan() {
    loading = true;
    fetchError = null;
    let growthPlan;
    try {
      growthPlan = await fetchGrowthPlan();
      if (paymentIssues?.length) {
        const returnUrl = getBillingUpgradeUrl($page, organization) + "?upgradeToGrowth=true";
        window.open(
          await fetchPaymentsPortalURL(
            organization,
            returnUrl,
          ),
          "_self",
        );
        return;
      }
    } catch (e) {
      loading = false;
      fetchError =
        e instanceof Error ? e.message : "An unexpected error occurred";
      return;
    }
    loading = false;

    if (type === "renew") {
      await $planRenewer.mutateAsync({
        org: organization,
        data: {
          planName: growthPlan?.name,
        },
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Your Growth plan was renewed",
      });
    } else {
      await $planUpdater.mutateAsync({
        org: organization,
        data: {
          planName: growthPlan?.name,
        },
      });
      showWelcomeToRillDialog.set(true);
    }
    void invalidateBillingInfo(organization);
    open = false;
    if (redirect) {
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
        <div>
          {description}
          <a
            href="https://www.rilldata.com/pricing"
            target="_blank"
            rel="noreferrer noopener"
          >
            See pricing details ->
          </a>

          <ul class="mt-5 ml-5 list-disc">
            <li>Pure usage-based billing, no base fee</li>
            <li>Managed: $0.15/slot/hr + $1/GB/month storage above 1GB</li>
            <li>Live Connect: Cluster Slots $0.06/hr + Rill Slots $0.15/hr</li>
          </ul>
        </div>
      </AlertDialogDescription>

      {#if $allStatus.isError || fetchError}
        <div class="text-red-500 text-sm py-px">
          {#if fetchError}
            <div>{fetchError}</div>
          {/if}
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
