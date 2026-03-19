<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    createExtendTrialMutation,
    createBillingRepairMutation,
    createDeleteBillingIssueMutation,
    createSetBillingCustomerMutation,
    getBillingIssues,
  } from "@rilldata/web-admin/features/admin/billing/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmAction: () => Promise<void> = async () => {};

  // Form state
  let trialOrg = "";
  let trialDays = 14;
  let repairOrg = "";
  let customerIdOrg = "";
  let customerId = "";
  let issuesOrg = "";

  const queryClient = useQueryClient();
  const extendTrial = createExtendTrialMutation();
  const billingRepair = createBillingRepairMutation();
  const deleteBillingIssue = createDeleteBillingIssueMutation();
  const setCustomer = createSetBillingCustomerMutation();

  $: billingIssuesQuery = getBillingIssues(issuesOrg);

  async function handleExtendTrial() {
    if (!trialOrg) return;
    try {
      await $extendTrial.mutateAsync({
        data: { org: trialOrg, days: trialDays },
      });
      bannerRef.show("success", `Trial extended by ${trialDays} days for ${trialOrg}`);
      trialOrg = "";
    } catch (err) {
      bannerRef.show("error", `Failed to extend trial: ${err}`);
    }
  }

  async function handleBillingRepair() {
    if (!repairOrg) return;
    confirmTitle = "Trigger Billing Repair";
    confirmDescription = `This will trigger a billing repair for organization "${repairOrg}". This recalculates billing state.`;
    confirmAction = async () => {
      try {
        await $billingRepair.mutateAsync({ data: { org: repairOrg } });
        bannerRef.show("success", `Billing repair triggered for ${repairOrg}`);
        repairOrg = "";
      } catch (err) {
        bannerRef.show("error", `Failed to trigger billing repair: ${err}`);
      }
    };
    confirmOpen = true;
  }

  async function handleSetCustomerId() {
    if (!customerIdOrg || !customerId) return;
    try {
      await $setCustomer.mutateAsync({
        data: { org: customerIdOrg, billingCustomerId: customerId },
      });
      bannerRef.show("success", `Billing customer ID set for ${customerIdOrg}`);
      customerIdOrg = "";
      customerId = "";
    } catch (err) {
      bannerRef.show("error", `Failed to set billing customer ID: ${err}`);
    }
  }

  async function handleDeleteIssue(org: string, type: string) {
    try {
      await $deleteBillingIssue.mutateAsync({ org, type });
      bannerRef.show("success", `Billing issue "${type}" deleted for ${org}`);
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed to delete billing issue: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Billing"
  description="Extend trials, repair billing state, manage billing customer IDs, and resolve billing issues."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <!-- Extend Trial -->
  <section class="card">
    <h2 class="card-title">Extend Trial</h2>
    <p class="card-desc">Add days to an organization's trial period.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={trialOrg}
      />
      <input
        type="number"
        class="input w-24"
        min="1"
        max="365"
        bind:value={trialDays}
      />
      <button class="btn-primary" on:click={handleExtendTrial}>
        Extend Trial
      </button>
    </div>
  </section>

  <!-- Set Billing Customer ID -->
  <section class="card">
    <h2 class="card-title">Set Billing Customer ID</h2>
    <p class="card-desc">Associate a Stripe customer ID with an organization.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={customerIdOrg}
      />
      <input
        type="text"
        class="input"
        placeholder="Stripe customer ID (cus_...)"
        bind:value={customerId}
      />
      <button class="btn-primary" on:click={handleSetCustomerId}>
        Set Customer ID
      </button>
    </div>
  </section>

  <!-- Billing Repair -->
  <section class="card">
    <h2 class="card-title">Billing Repair</h2>
    <p class="card-desc">Trigger a billing state recalculation for an organization.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={repairOrg}
      />
      <button class="btn-primary" on:click={handleBillingRepair}>
        Trigger Repair
      </button>
    </div>
  </section>

  <!-- Billing Issues -->
  <section class="card">
    <h2 class="card-title">Billing Issues</h2>
    <p class="card-desc">View and resolve billing issues for an organization.</p>
    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={issuesOrg}
      />
    </div>
    {#if $billingIssuesQuery.data?.issues?.length}
      <div class="issues-list">
        {#each $billingIssuesQuery.data.issues as issue}
          <div class="issue-row">
            <div>
              <span class="issue-type">{issue.type}</span>
              <span class="issue-meta">{issue.metadata ?? ""}</span>
            </div>
            <button
              class="action-btn destructive"
              on:click={() => handleDeleteIssue(issuesOrg, issue.type ?? "")}
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
    @apply flex flex-col gap-6;
  }

  .card {
    @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700;
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
      hover:bg-blue-700 whitespace-nowrap;
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
</style>
