<script lang="ts">
  import SuperuserPageHeader from "@rilldata/web-admin/features/superuser/layout/SuperuserPageHeader.svelte";
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { V1BillingIssueType } from "@rilldata/web-admin/client";
  import {
    getBillingSetupURL,
    createExtendTrialMutation,
    createDeleteBillingIssueMutation,
    getBillingIssues,
  } from "@rilldata/web-admin/features/superuser/billing/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  // Billing Setup state
  let setupOrg = "";
  let setupLoading = false;
  let setupUrl = "";

  // Extend Trial state
  let trialOrg = "";
  let trialDays = 14;
  let trialLoading = false;

  // Billing Issues state
  let issuesOrg = "";

  // Confirmation dialog state
  let dialogOpen = false;
  let dialogTitle = "";
  let dialogDescription = "";
  let dialogDestructive = false;
  let dialogAction: () => Promise<void> = async () => {};
  let dialogLoading = false;

  const queryClient = useQueryClient();
  const extendTrial = createExtendTrialMutation();
  const deleteBillingIssue = createDeleteBillingIssueMutation();

  $: billingIssuesQuery = getBillingIssues(issuesOrg);

  async function handleBillingSetup() {
    if (!setupOrg) return;
    setupLoading = true;
    setupUrl = "";
    try {
      const url = await getBillingSetupURL(setupOrg);
      if (url) {
        setupUrl = url;
        eventBus.emit("notification", {
          type: "success",
          message: `Billing setup URL generated for ${setupOrg}`,
        });
      } else {
        eventBus.emit("notification", {
          type: "error",
          message: "No URL returned; check the org name.",
        });
      }
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to generate billing setup URL: ${err}`,
      });
    } finally {
      setupLoading = false;
    }
  }

  function confirmExtendTrial() {
    if (!trialOrg) return;
    dialogTitle = "Extend Trial";
    dialogDescription = `This will extend the trial for "${trialOrg}" by ${trialDays} day${trialDays === 1 ? "" : "s"}.`;
    dialogDestructive = false;
    dialogAction = async () => {
      trialLoading = true;
      try {
        await $extendTrial.mutateAsync({
          data: { org: trialOrg, days: trialDays },
        });
        eventBus.emit("notification", {
          type: "success",
          message: `Trial extended by ${trialDays} days for ${trialOrg}`,
        });
        trialOrg = "";
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to extend trial: ${err}`,
        });
      } finally {
        trialLoading = false;
      }
    };
    dialogOpen = true;
  }

  function confirmDeleteIssue(org: string, type: V1BillingIssueType) {
    dialogTitle = "Delete Billing Issue";
    dialogDescription = `This will delete the billing issue "${type}" for "${org}". This action cannot be undone.`;
    dialogDestructive = true;
    dialogAction = async () => {
      try {
        await $deleteBillingIssue.mutateAsync({ org, type });
        eventBus.emit("notification", {
          type: "success",
          message: `Billing issue "${type}" deleted for ${org}`,
        });
        await queryClient.invalidateQueries({
          predicate: (q) =>
            (q.queryKey[0] as string)?.includes("/v1/organizations") ||
            (q.queryKey[0] as string)?.includes("/v1/superuser/billing"),
        });
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to delete billing issue: ${err}`,
        });
      }
    };
    dialogOpen = true;
  }

  async function handleConfirm() {
    dialogLoading = true;
    try {
      await dialogAction();
      dialogOpen = false;
    } catch {
      // Keep open for retry
    } finally {
      dialogLoading = false;
    }
  }
</script>

<SuperuserPageHeader
  title="Billing"
  description="Generate billing setup links, extend trials, and manage billing issues."
/>

<div class="flex flex-col gap-6 pb-12">
  <!-- Billing Setup -->
  <section class="p-5 rounded-lg border border-primary-200 bg-primary-50/50">
    <h2 class="text-sm font-semibold text-fg-primary mb-1">Billing Setup</h2>
    <p class="text-sm text-fg-secondary mb-4">
      Generate a Stripe checkout page link for an organization to enter their
      billing information.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgPicker bind:value={setupOrg} />
      </div>
      <Button
        large
        class="font-normal"
        type="primary"
        onClick={handleBillingSetup}
        disabled={setupLoading || !setupOrg}
        loading={setupLoading}
      >
        Generate Setup Link
      </Button>
    </div>
    {#if setupUrl}
      <div class="mt-4 flex flex-col gap-1">
        <span class="text-sm text-fg-secondary"
          >Share this link with the customer:</span
        >
        <div
          class="flex items-center gap-2 p-3 rounded-md bg-surface-subtle border"
        >
          <a
            href={setupUrl}
            target="_blank"
            rel="noreferrer"
            class="flex-1 text-sm text-accent-primary-action break-all hover:underline"
          >
            {setupUrl}
          </a>
          <Button
            large
            class="font-normal"
            type="tertiary"
            onClick={() => {
              navigator.clipboard.writeText(setupUrl);
              eventBus.emit("notification", {
                type: "success",
                message: "URL copied to clipboard",
              });
            }}
          >
            Copy
          </Button>
        </div>
      </div>
    {/if}
  </section>

  <!-- Extend Trial -->
  <section class="p-5 rounded-lg border">
    <h2 class="text-sm font-semibold text-fg-primary mb-1">Extend Trial</h2>
    <p class="text-sm text-fg-secondary mb-4">
      Add days to an organization's trial period.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgPicker bind:value={trialOrg} />
      </div>
      <input
        type="number"
        class="w-24 px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        min="1"
        max="30"
        bind:value={trialDays}
      />
      <Button
        large
        class="font-normal"
        type="primary"
        onClick={confirmExtendTrial}
        disabled={trialLoading || !trialOrg}
        loading={trialLoading}
      >
        Extend Trial
      </Button>
    </div>
  </section>

  <!-- Billing Issues -->
  <section class="p-5 rounded-lg border">
    <h2 class="text-sm font-semibold text-fg-primary mb-1">Billing Issues</h2>
    <p class="text-sm text-fg-secondary mb-4">
      View and resolve billing issues for an organization.
    </p>
    <div class="flex gap-3 items-center flex-wrap mb-4">
      <div class="w-64">
        <OrgPicker bind:value={issuesOrg} />
      </div>
    </div>
    {#if $billingIssuesQuery.isFetching}
      <p class="text-sm text-fg-secondary py-2">Loading issues...</p>
    {:else if $billingIssuesQuery.data?.issues?.length}
      <div class="flex flex-col gap-2">
        {#each $billingIssuesQuery.data.issues as issue}
          <div
            class="flex items-center justify-between px-3 py-2 rounded bg-surface-subtle"
          >
            <div>
              <span class="text-sm font-mono text-fg-primary">{issue.type}</span
              >
              <span class="text-sm text-fg-secondary ml-2"
                >{issue.metadata ?? ""}</span
              >
            </div>
            <Button
              large
              class="font-normal"
              type="secondary-destructive"
              onClick={() =>
                confirmDeleteIssue(
                  issuesOrg,
                  issue.type ?? "BILLING_ISSUE_TYPE_UNSPECIFIED",
                )}
            >
              Delete Issue
            </Button>
          </div>
        {/each}
      </div>
    {:else if issuesOrg && $billingIssuesQuery.isSuccess}
      <p class="text-sm text-fg-secondary">No billing issues found.</p>
    {/if}
  </section>
</div>

<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{dialogTitle}</AlertDialogTitle>
      <AlertDialogDescription>{dialogDescription}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        large
        class="font-normal"
        type="tertiary"
        onClick={() => (dialogOpen = false)}>Cancel</Button
      >
      <Button
        large
        class="font-normal"
        type={dialogDestructive ? "destructive" : "primary"}
        onClick={handleConfirm}
        loading={dialogLoading}
      >
        Confirm
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
