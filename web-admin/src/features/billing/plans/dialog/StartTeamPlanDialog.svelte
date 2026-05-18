<script lang="ts">
  import { page } from "$app/state";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types.ts";
  import { getSubscriptionResumedText } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
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
  import { upgradeToPro } from "@rilldata/web-admin/features/billing/plans/upgrade-to-pro.ts";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors.ts";

  let {
    open = $bindable(false),
    type,
    organization,
    endDate = "",
  }: {
    open: boolean;
    type: TeamPlanDialogTypes;
    organization: string;
    endDate?: string;
  } = $props();

  let title: string = $state("");
  let description =
    $state(`Your subscription will start today using the payment method on file.
    You'll be billed monthly based on usage at $0.15/compute unit/hr and $1/GB storage/mo.
    Cancel anytime.`);
  let buttonText = $state("Upgrade to Pro");
  function setCopyBasedOnType(t: TeamPlanDialogTypes) {
    switch (t) {
      case "trial-expired": // No explicit messaging for this as of now
      case "base":
        title = "Upgrade to Pro";
        buttonText = "Continue";
        break;

      case "size":
        title = "Deploying more than 10GB requires a Pro plan";
        break;

      case "org":
        title = "To create another organization, start a Pro plan";
        description = "";
        break;

      case "proj":
        title = "To deploy a second project, start a Pro plan";
        break;

      case "renew":
        title = "Renew Pro plan";
        description = `Your billing cycle will resume ${getSubscriptionResumedText(endDate)}.`;
        buttonText = "Continue";
        break;
    }
  }
  $effect(() => setCopyBasedOnType(type));

  let categorisedIssuesQuery = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let categorisedIssues = $derived($categorisedIssuesQuery.data);
  let redirect = $derived(page.url.searchParams.get("redirect"));

  let loading = $state(false);
  let fetchError = $state<string | null>(null);

  async function handleUpgradePlan() {
    loading = true;
    fetchError = null;
    try {
      await upgradeToPro(organization, categorisedIssues, redirect);
    } catch (e) {
      fetchError = extractErrorMessage(e);
    }
    loading = false;
    open = false;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>

      <AlertDialogDescription>
        {description}
      </AlertDialogDescription>

      {#if fetchError}
        <div class="text-red-500 text-sm py-px">
          {#if fetchError}
            <div>{fetchError}</div>
          {/if}
        </div>
      {/if}
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button type="secondary" onClick={() => (open = false)}>Close</Button>
      <Button type="primary" onClick={handleUpgradePlan} {loading}>
        {buttonText}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
