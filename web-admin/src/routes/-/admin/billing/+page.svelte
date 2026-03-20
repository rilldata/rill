<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import OrgSearchInput from "@rilldata/web-admin/features/admin/shared/OrgSearchInput.svelte";
  import type { V1BillingIssueType } from "@rilldata/web-admin/client";
  import {
    getBillingSetupURL,
    createExtendTrialMutation,
    createBillingRepairMutation,
    createDeleteBillingIssueMutation,
    createSetBillingCustomerMutation,
    getBillingIssues,
  } from "@rilldata/web-admin/features/admin/billing/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmAction: () => Promise<void> = async () => {};

  // Billing Setup state
  let setupOrg = "";
  let setupLoading = false;
  let setupUrl = "";

  // Form state
  let trialOrg = "";
  let trialDays = 14;
  let trialLoading = false;
  let repairOrg = "";
  let customerIdOrg = "";
  let customerId = "";
  let customerIdLoading = false;
  let issuesOrg = "";

  const queryClient = useQueryClient();
  const extendTrial = createExtendTrialMutation();
  const billingRepair = createBillingRepairMutation();
  const deleteBillingIssue = createDeleteBillingIssueMutation();
  const setCustomer = createSetBillingCustomerMutation();

  $: billingIssuesQuery = getBillingIssues(issuesOrg);

  async function handleBillingSetup() {
    if (!setupOrg) return;
    setupLoading = true;
    setupUrl = "";
    try {
      const url = await getBillingSetupURL(setupOrg);
      if (url) {
        setupUrl = url;
        notifySuccess( `Billing setup URL generated for ${setupOrg}`);
      } else {
        notifyError( "No URL returned; check the org name.");
      }
    } catch (err) {
      notifyError( `Failed to generate billing setup URL: ${err}`);
    } finally {
      setupLoading = false;
    }
  }

  async function handleExtendTrial() {
    if (!trialOrg) return;
    trialLoading = true;
    try {
      await $extendTrial.mutateAsync({
        data: { org: trialOrg, days: trialDays },
      });
      notifySuccess(
        `Trial extended by ${trialDays} days for ${trialOrg}`,
      );
      trialOrg = "";
    } catch (err) {
      notifyError( `Failed to extend trial: ${err}`);
    } finally {
      trialLoading = false;
    }
  }

  async function handleBillingRepair() {
    if (!repairOrg) return;
    confirmTitle = "Trigger Billing Repair";
    confirmDescription = `This will trigger a billing repair for organization "${repairOrg}". This recalculates billing state.`;
    confirmAction = async () => {
      try {
        await $billingRepair.mutateAsync({ data: { org: repairOrg } });
        notifySuccess( `Billing repair triggered for ${repairOrg}`);
        repairOrg = "";
      } catch (err) {
        notifyError( `Failed to trigger billing repair: ${err}`);
      }
    };
    confirmOpen = true;
  }

  async function handleSetCustomerId() {
    if (!customerIdOrg || !customerId) return;
    customerIdLoading = true;
    try {
      await $setCustomer.mutateAsync({
        data: { org: customerIdOrg, billingCustomerId: customerId },
      });
      notifySuccess(
        `Billing customer ID set for ${customerIdOrg}`,
      );
      customerIdOrg = "";
      customerId = "";
    } catch (err) {
      notifyError( `Failed to set billing customer ID: ${err}`);
    } finally {
      customerIdLoading = false;
    }
  }

  async function handleDeleteIssue(org: string, type: V1BillingIssueType) {
    try {
      await $deleteBillingIssue.mutateAsync({ org, type });
      notifySuccess( `Billing issue "${type}" deleted for ${org}`);
      await queryClient.invalidateQueries({
        predicate: (q) =>
          (q.queryKey[0] as string)?.includes("/v1/organizations") ||
          (q.queryKey[0] as string)?.includes("/v1/superuser/billing"),
      });
    } catch (err) {
      notifyError( `Failed to delete billing issue: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Billing"
  description="Generate billing setup links, extend trials, repair billing state, and manage billing issues."
/>

<div class="sections">
  <!-- Billing Setup (Primary feature) -->
  <section class="card highlight">
    <h2 class="card-title">Billing Setup</h2>
    <p class="card-desc">
      Generate a Stripe checkout page link for an organization to enter their
      billing information.
    </p>
    <div class="form-row">
      <div class="w-64">
        <OrgSearchInput
          bind:value={setupOrg}
          placeholder="Search organization..."
        />
      </div>
      <button
        class="btn-primary"
        on:click={handleBillingSetup}
        disabled={setupLoading || !setupOrg}
      >
        {#if setupLoading}
          <span class="btn-spinner" />
          Generating...
        {:else}
          Generate Setup Link
        {/if}
      </button>
    </div>
    {#if setupUrl}
      <div class="url-result">
        <span class="text-xs text-slate-500"
          >Share this link with the customer:</span
        >
        <div class="url-box">
          <a
            href={setupUrl}
            target="_blank"
            rel="noreferrer"
            class="url-link"
          >
            {setupUrl}
          </a>
          <button
            class="copy-btn"
            on:click={() => {
              navigator.clipboard.writeText(setupUrl);
              notifySuccess( "URL copied to clipboard");
            }}
          >
            Copy
          </button>
        </div>
      </div>
    {/if}
  </section>

  <!-- Extend Trial -->
  <section class="card">
    <h2 class="card-title">Extend Trial</h2>
    <p class="card-desc">Add days to an organization's trial period.</p>
    <div class="form-row">
      <div class="w-64">
        <OrgSearchInput
          bind:value={trialOrg}
          placeholder="Search organization..."
        />
      </div>
      <input
        type="number"
        class="input w-24"
        min="1"
        max="365"
        bind:value={trialDays}
      />
      <button
        class="btn-primary"
        on:click={handleExtendTrial}
        disabled={trialLoading || !trialOrg}
      >
        {#if trialLoading}
          <span class="btn-spinner" />
          Extending...
        {:else}
          Extend Trial
        {/if}
      </button>
    </div>
  </section>

  <!-- Set Billing Customer ID -->
  <section class="card">
    <h2 class="card-title">Set Billing Customer ID</h2>
    <p class="card-desc">
      Associate a Stripe customer ID with an organization.
    </p>
    <div class="form-row">
      <div class="w-64">
        <OrgSearchInput
          bind:value={customerIdOrg}
          placeholder="Search organization..."
        />
      </div>
      <input
        type="text"
        class="input"
        placeholder="Stripe customer ID (cus_...)"
        bind:value={customerId}
      />
      <button
        class="btn-primary"
        on:click={handleSetCustomerId}
        disabled={customerIdLoading || !customerIdOrg || !customerId}
      >
        {#if customerIdLoading}
          <span class="btn-spinner" />
          Setting...
        {:else}
          Set Customer ID
        {/if}
      </button>
    </div>
  </section>

  <!-- Billing Repair -->
  <section class="card">
    <h2 class="card-title">Billing Repair</h2>
    <p class="card-desc">
      Trigger a billing state recalculation for an organization.
    </p>
    <div class="form-row">
      <div class="w-64">
        <OrgSearchInput
          bind:value={repairOrg}
          placeholder="Search organization..."
        />
      </div>
      <button
        class="btn-primary"
        on:click={handleBillingRepair}
        disabled={!repairOrg}
      >
        Trigger Repair
      </button>
    </div>
  </section>

  <!-- Billing Issues -->
  <section class="card">
    <h2 class="card-title">Billing Issues</h2>
    <p class="card-desc">
      View and resolve billing issues for an organization.
    </p>
    <div class="form-row mb-4">
      <div class="w-64">
        <OrgSearchInput
          bind:value={issuesOrg}
          placeholder="Search organization..."
        />
      </div>
    </div>
    {#if $billingIssuesQuery.isFetching}
      <div class="loading">
        <div class="spinner" />
        <span class="text-sm text-slate-500">Loading issues...</span>
      </div>
    {:else if $billingIssuesQuery.data?.issues?.length}
      <div class="issues-list">
        {#each $billingIssuesQuery.data.issues as issue}
          <div class="issue-row">
            <div>
              <span class="issue-type">{issue.type}</span>
              <span class="issue-meta">{issue.metadata ?? ""}</span>
            </div>
            <button
              class="action-btn destructive"
              on:click={() =>
                handleDeleteIssue(issuesOrg, issue.type ?? "BILLING_ISSUE_TYPE_UNSPECIFIED")}
            >
              Delete Issue
            </button>
          </div>
        {/each}
      </div>
    {:else if issuesOrg && $billingIssuesQuery.isSuccess}
      <p class="text-sm text-slate-500">No billing issues found.</p>
    {/if}
  </section>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  .sections {
    @apply flex flex-col gap-6 pb-12;
  }

  .card {
    @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700;
  }

  .card.highlight {
    @apply border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-900/10;
  }

  .card-title {
    @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-1;
  }

  .card-desc {
    @apply text-xs text-slate-500 dark:text-slate-400 mb-4;
  }

  .form-row {
    @apply flex gap-3 items-center flex-wrap;
  }

  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300
      dark:border-slate-600 bg-white dark:bg-slate-800
      text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 dark:placeholder:text-slate-500
      focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }

  .btn-primary {
    @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white
      hover:bg-blue-700 whitespace-nowrap flex items-center gap-2;
  }

  .btn-primary:disabled {
    @apply opacity-50 cursor-not-allowed;
  }

  .btn-spinner {
    @apply inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin;
  }

  .url-result {
    @apply mt-4 flex flex-col gap-1;
  }

  .url-box {
    @apply flex items-center gap-2 p-3 rounded-md bg-white dark:bg-slate-800
      border border-slate-200 dark:border-slate-700;
  }

  .url-link {
    @apply flex-1 text-sm text-blue-600 dark:text-blue-400 break-all hover:underline;
  }

  .copy-btn {
    @apply text-xs px-3 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700
      whitespace-nowrap;
  }

  .issues-list {
    @apply flex flex-col gap-2;
  }

  .issue-row {
    @apply flex items-center justify-between px-3 py-2 rounded
      bg-slate-50 dark:bg-slate-800;
  }

  .issue-type {
    @apply text-sm font-mono text-slate-700 dark:text-slate-300;
  }

  .issue-meta {
    @apply text-xs text-slate-500 ml-2;
  }

  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300;
  }

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }

  .loading {
    @apply flex items-center gap-2 py-2;
  }

  .spinner {
    @apply w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin;
  }
</style>
