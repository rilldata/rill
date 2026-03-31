<script lang="ts">
  import OrgPicker from "@rilldata/web-admin/features/superuser/shared/OrgPicker.svelte";
  import ConfirmActionDialog from "@rilldata/web-admin/features/superuser/dialogs/ConfirmActionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    type V1BillingIssueType,
    getAdminServiceListOrganizationBillingIssuesQueryKey,
  } from "@rilldata/web-admin/client";
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

  // Extend Trial dialog state
  let extendDialogOpen = false;

  // Billing Issues state
  let issuesOrg = "";

  // Delete Billing Issue dialog state
  let deleteIssueDialogOpen = false;
  let deleteIssueOrg = "";
  let deleteIssueType: V1BillingIssueType = "BILLING_ISSUE_TYPE_UNSPECIFIED";

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

  async function doExtendTrial() {
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
      throw err;
    } finally {
      trialLoading = false;
    }
  }

  async function doDeleteIssue() {
    try {
      await $deleteBillingIssue.mutateAsync({
        org: deleteIssueOrg,
        type: deleteIssueType,
      });
      eventBus.emit("notification", {
        type: "success",
        message: `Billing issue "${deleteIssueType}" deleted for ${deleteIssueOrg}`,
      });
      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationBillingIssuesQueryKey(deleteIssueOrg),
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to delete billing issue: ${err}`,
      });
      throw err;
    }
  }
</script>

<p class="text-sm text-fg-secondary mb-4">
  Generate billing setup links, extend trials, and manage billing issues.
</p>

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
        onClick={() => {
          if (trialOrg) extendDialogOpen = true;
        }}
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
              onClick={() => {
                deleteIssueOrg = issuesOrg;
                deleteIssueType =
                  issue.type ?? "BILLING_ISSUE_TYPE_UNSPECIFIED";
                deleteIssueDialogOpen = true;
              }}
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

<ConfirmActionDialog
  bind:open={extendDialogOpen}
  title="Extend Trial"
  description={`This will extend the trial for "${trialOrg}" by ${trialDays} day${trialDays === 1 ? "" : "s"}.`}
  onConfirm={doExtendTrial}
/>

<ConfirmActionDialog
  bind:open={deleteIssueDialogOpen}
  title="Delete Billing Issue"
  description={`This will delete the billing issue "${deleteIssueType}" for "${deleteIssueOrg}". This action cannot be undone.`}
  confirmLabel="Delete"
  onConfirm={doDeleteIssue}
/>
