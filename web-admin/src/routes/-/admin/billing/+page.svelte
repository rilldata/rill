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

<div class="flex flex-col gap-6 pb-12">
  <!-- Billing Setup (Primary feature) -->
  <section
    class="p-5 rounded-lg border border-blue-200 bg-blue-50/50"
  >
    <h2 class="text-sm font-semibold text-slate-900 mb-1">
      Billing Setup
    </h2>
    <p class="text-xs text-slate-500 mb-4">
      Generate a Stripe checkout page link for an organization to enter their
      billing information.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgSearchInput
          bind:value={setupOrg}
          placeholder="Search organization..."
        />
      </div>
      <button
        class="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={handleBillingSetup}
        disabled={setupLoading || !setupOrg}
      >
        {#if setupLoading}
          <span
            class="inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"
          />
          Generating...
        {:else}
          Generate Setup Link
        {/if}
      </button>
    </div>
    {#if setupUrl}
      <div class="mt-4 flex flex-col gap-1">
        <span class="text-xs text-slate-500"
          >Share this link with the customer:</span
        >
        <div
          class="flex items-center gap-2 p-3 rounded-md bg-slate-50 border border-slate-200"
        >
          <a
            href={setupUrl}
            target="_blank"
            rel="noreferrer"
            class="flex-1 text-sm text-blue-600 break-all hover:underline"
          >
            {setupUrl}
          </a>
          <button
            class="text-xs px-3 py-1 rounded border border-slate-300 text-slate-600 hover:bg-slate-100 whitespace-nowrap"
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
  <section class="p-5 rounded-lg border border-slate-200">
    <h2 class="text-sm font-semibold text-slate-900 mb-1">
      Extend Trial
    </h2>
    <p class="text-xs text-slate-500 mb-4">
      Add days to an organization's trial period.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgSearchInput
          bind:value={trialOrg}
          placeholder="Search organization..."
        />
      </div>
      <input
        type="number"
        class="w-24 px-3 py-2 text-sm rounded-md border border-slate-300 bg-slate-50 text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        min="1"
        max="365"
        bind:value={trialDays}
      />
      <button
        class="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={handleExtendTrial}
        disabled={trialLoading || !trialOrg}
      >
        {#if trialLoading}
          <span
            class="inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"
          />
          Extending...
        {:else}
          Extend Trial
        {/if}
      </button>
    </div>
  </section>

  <!-- Set Billing Customer ID -->
  <section class="p-5 rounded-lg border border-slate-200">
    <h2 class="text-sm font-semibold text-slate-900 mb-1">
      Set Billing Customer ID
    </h2>
    <p class="text-xs text-slate-500 mb-4">
      Associate a Stripe customer ID with an organization.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgSearchInput
          bind:value={customerIdOrg}
          placeholder="Search organization..."
        />
      </div>
      <input
        type="text"
        class="px-3 py-2 text-sm rounded-md border border-slate-300 bg-slate-50 text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        placeholder="Stripe customer ID (cus_...)"
        bind:value={customerId}
      />
      <button
        class="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={handleSetCustomerId}
        disabled={customerIdLoading || !customerIdOrg || !customerId}
      >
        {#if customerIdLoading}
          <span
            class="inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"
          />
          Setting...
        {:else}
          Set Customer ID
        {/if}
      </button>
    </div>
  </section>

  <!-- Billing Repair -->
  <section class="p-5 rounded-lg border border-slate-200">
    <h2 class="text-sm font-semibold text-slate-900 mb-1">
      Billing Repair
    </h2>
    <p class="text-xs text-slate-500 mb-4">
      Trigger a billing state recalculation for an organization.
    </p>
    <div class="flex gap-3 items-center flex-wrap">
      <div class="w-64">
        <OrgSearchInput
          bind:value={repairOrg}
          placeholder="Search organization..."
        />
      </div>
      <button
        class="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={handleBillingRepair}
        disabled={!repairOrg}
      >
        Trigger Repair
      </button>
    </div>
  </section>

  <!-- Billing Issues -->
  <section class="p-5 rounded-lg border border-slate-200">
    <h2 class="text-sm font-semibold text-slate-900 mb-1">
      Billing Issues
    </h2>
    <p class="text-xs text-slate-500 mb-4">
      View and resolve billing issues for an organization.
    </p>
    <div class="flex gap-3 items-center flex-wrap mb-4">
      <div class="w-64">
        <OrgSearchInput
          bind:value={issuesOrg}
          placeholder="Search organization..."
        />
      </div>
    </div>
    {#if $billingIssuesQuery.isFetching}
      <div class="flex items-center gap-2 py-2">
        <div
          class="w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
        />
        <span class="text-sm text-slate-500">Loading issues...</span>
      </div>
    {:else if $billingIssuesQuery.data?.issues?.length}
      <div class="flex flex-col gap-2">
        {#each $billingIssuesQuery.data.issues as issue}
          <div
            class="flex items-center justify-between px-3 py-2 rounded bg-slate-50"
          >
            <div>
              <span class="text-sm font-mono text-slate-700"
                >{issue.type}</span
              >
              <span class="text-xs text-slate-500 ml-2"
                >{issue.metadata ?? ""}</span
              >
            </div>
            <button
              class="text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50"
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
