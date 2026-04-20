<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    createAdminServiceGetBillingSubscription,
  } from "@rilldata/web-admin/client";
  import {
    fetchPaymentsPortalURL,
    getBillingUpgradeUrl,
    getOrganizationUsageMetrics,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { page } from "$app/stores";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
  let organization = $derived(data.organization);

  // Current deployment data for pre-filling
  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization),
  );
  let projects = $derived($projectsQuery.data?.projects ?? []);

  let currentProdSlots = $derived(
    projects.reduce((sum, p) => sum + Number(p.prodSlots ?? 0), 0),
  );
  let currentDevSlots = $derived(
    projects.reduce((sum, p) => sum + Number(p.devSlots ?? 0), 0),
  );

  let usageMetrics = $derived(getOrganizationUsageMetrics(organization));
  let currentStorageBytes = $derived(
    $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0,
  );
  let currentStorageGB = $derived(Math.ceil(currentStorageBytes / 1e9) || 1);

  // Editable inputs
  let prodUnits = $state(2);
  let prodHoursPerDay = $state(24);
  let devUnits = $state(0);
  let devHoursPerDay = $state(24);
  let storageGB = $state(1);

  // Pre-fill from current deployment once data loads
  let prefilled = $state(false);
  $effect(() => {
    if (!prefilled && !$projectsQuery.isLoading && projects.length > 0) {
      prodUnits = Math.max(currentProdSlots, 2);
      devUnits = currentDevSlots;
      storageGB = Math.max(currentStorageGB, 1);
      prefilled = true;
    }
  });

  function resetToUsage() {
    prodUnits = Math.max(currentProdSlots, 2);
    prodHoursPerDay = 24;
    devUnits = currentDevSlots;
    devHoursPerDay = 24;
    storageGB = Math.max(currentStorageGB, 1);
  }

  // Pricing constants
  const RATE_PER_UNIT_HR = 0.15;
  const STORAGE_RATE_PER_GB = 1;
  const FREE_STORAGE_GB = 1;
  const DAYS_PER_MONTH = 30;

  // Cost calculations
  let prodCost = $derived(
    prodUnits * prodHoursPerDay * DAYS_PER_MONTH * RATE_PER_UNIT_HR,
  );
  let devCost = $derived(
    devUnits * devHoursPerDay * DAYS_PER_MONTH * RATE_PER_UNIT_HR,
  );
  let billableStorageGB = $derived(Math.max(storageGB - FREE_STORAGE_GB, 0));
  let storageCost = $derived(billableStorageGB * STORAGE_RATE_PER_GB);
  let monthlyCost = $derived(prodCost + devCost + storageCost);
  let dailyCost = $derived(monthlyCost / DAYS_PER_MONTH);

  // TODO: Wire to billing API when available
  let availableCredit = $derived(100);
  let firstBill = $derived(Math.max(monthlyCost - availableCredit, 0));

  // Billing plan
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isPro = $derived(planName === "pro_plan");

  function fmtUSD(n: number): string {
    return n.toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    });
  }

  let isCurrentConfig = $derived(
    prodUnits === Math.max(currentProdSlots, 2) &&
      prodHoursPerDay === 24 &&
      devUnits === currentDevSlots &&
      devHoursPerDay === 24 &&
      storageGB === Math.max(currentStorageGB, 1),
  );

  async function handleSubscribe() {
    window.open(
      await fetchPaymentsPortalURL(
        organization,
        getBillingUpgradeUrl($page, organization),
        true,
      ),
      "_self",
    );
  }

  function handleContactSales() {
    window.Pylon("show");
  }

  function clampInput(
    e: Event,
    min: number,
    max: number,
    setter: (v: number) => void,
  ) {
    const el = e.target as HTMLInputElement;
    const n = parseInt(el.value, 10);
    if (isNaN(n) || n < min) setter(min);
    else if (n > max) setter(max);
    else setter(n);
  }
</script>

<div class="estimate-page">
  <!-- Back link + title -->
  <a href="/{organization}/-/settings/billing" class="back-link">
    <svg
      class="w-3.5 h-3.5"
      viewBox="0 0 14 14"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
    >
      <path d="M9 2.5L4.5 7 9 11.5" />
    </svg>
    Billing
  </a>
  <h1 class="page-title">Estimate your cost</h1>

  <div class="page-header">
    <span class="prefill-note">Pre-filled from your current deployment</span>
    <button class="reset-btn" onclick={resetToUsage} disabled={isCurrentConfig}>
      <svg
        class="w-3.5 h-3.5"
        viewBox="0 0 14 14"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path d="M1.5 2.5v3.5h3.5" />
        <path d="M2.1 8.5a5 5 0 1 0 .9-4.5L1.5 6" />
      </svg>
      Reset to current configuration
    </button>
  </div>

  <div class="estimate-grid">
    <!-- LEFT: Input panel -->
    <div class="input-panel">
      <div class="panel-header">
        <span class="panel-title">Select compute unit</span>
        <span class="unit-hint">1 unit = 4 GiB RAM, 1 vCPU · $0.15/unit/hr</span>
      </div>

      <!-- Production units -->
      <div class="input-row">
        <div class="input-info">
          <span class="input-label">Production</span>
          <span class="input-desc">Minimum 2 units · $0.15/unit/hr</span>
        </div>
        <div class="stepper">
          <button
            class="stepper-btn"
            onclick={() => prodUnits--}
            disabled={prodUnits <= 2}>−</button
          >
          <input
            class="stepper-input"
            type="number"
            bind:value={prodUnits}
            min="2"
            onblur={(e) => clampInput(e, 2, 9999, (v) => (prodUnits = v))}
          />
          <button class="stepper-btn" onclick={() => prodUnits++}>+</button>
        </div>
      </div>

      <!-- Production active hours -->
      <div class="input-row input-row-sub">
        <div class="input-info">
          <span class="input-label-sub inline-flex items-center gap-1">
            Active hours per day
            <Tooltip location="right" alignment="middle" distance={8}>
              <span class="text-fg-muted flex">
                <InfoCircle size="13px" />
              </span>
              <TooltipContent maxWidth="220px" slot="tooltip-content">
                We hibernate your deployment when it's inactive, saving you on
                cost.
              </TooltipContent>
            </Tooltip>
          </span>
        </div>
        <div class="stepper">
          <button
            class="stepper-btn"
            onclick={() => prodHoursPerDay--}
            disabled={prodHoursPerDay <= 1}>−</button
          >
          <input
            class="stepper-input"
            type="number"
            bind:value={prodHoursPerDay}
            min="1"
            max="24"
            onblur={(e) =>
              clampInput(e, 1, 24, (v) => (prodHoursPerDay = v))}
          />
          <button
            class="stepper-btn"
            onclick={() => prodHoursPerDay++}
            disabled={prodHoursPerDay >= 24}>+</button
          >
        </div>
      </div>

      <div class="input-divider"></div>

      <!-- Development units -->
      <div class="input-row">
        <div class="input-info">
          <span class="input-label">Development</span>
          <span class="input-desc">Same rate as prod unit</span>
        </div>
        <div class="stepper">
          <button
            class="stepper-btn"
            onclick={() => devUnits--}
            disabled={devUnits <= 0}>−</button
          >
          <input
            class="stepper-input"
            type="number"
            bind:value={devUnits}
            min="0"
            onblur={(e) => clampInput(e, 0, 9999, (v) => (devUnits = v))}
          />
          <button class="stepper-btn" onclick={() => devUnits++}>+</button>
        </div>
      </div>

      <!-- Development active hours -->
      <div class="input-row input-row-sub">
        <div class="input-info">
          <span class="input-label-sub inline-flex items-center gap-1">
            Active hours per day
            <Tooltip location="right" alignment="middle" distance={8}>
              <span class="text-fg-muted flex">
                <InfoCircle size="13px" />
              </span>
              <TooltipContent maxWidth="220px" slot="tooltip-content">
                We hibernate your deployment when it's inactive, saving you on
                cost.
              </TooltipContent>
            </Tooltip>
          </span>
        </div>
        <div class="stepper">
          <button
            class="stepper-btn"
            onclick={() => devHoursPerDay--}
            disabled={devHoursPerDay <= 1}>−</button
          >
          <input
            class="stepper-input"
            type="number"
            bind:value={devHoursPerDay}
            min="1"
            max="24"
            onblur={(e) => clampInput(e, 1, 24, (v) => (devHoursPerDay = v))}
          />
          <button
            class="stepper-btn"
            onclick={() => devHoursPerDay++}
            disabled={devHoursPerDay >= 24}>+</button
          >
        </div>
      </div>

      <div class="input-divider"></div>

      <!-- Storage -->
      <div class="input-row">
        <div class="input-info">
          <span class="input-label">Storage (GB)</span>
          <span class="input-desc"
            >1 GB included free · $1/GB/mo above that</span
          >
        </div>
        <div class="stepper">
          <button
            class="stepper-btn"
            onclick={() => storageGB--}
            disabled={storageGB <= 1}>−</button
          >
          <input
            class="stepper-input"
            type="number"
            bind:value={storageGB}
            min="1"
            onblur={(e) => clampInput(e, 1, 9999, (v) => (storageGB = v))}
          />
          <button class="stepper-btn" onclick={() => storageGB++}>+</button>
        </div>
      </div>
    </div>

    <!-- RIGHT: Cost summary -->
    <div class="cost-panel">
      <h2 class="cost-title">Estimate monthly cost</h2>
      <span class="cost-total">{fmtUSD(monthlyCost)}</span>
      <span class="cost-daily">~{fmtUSD(dailyCost)}/day</span>

      <div class="cost-divider"></div>

      <!-- Production breakdown -->
      <div class="cost-row">
        <div class="cost-row-info">
          <span class="cost-row-label">Production</span>
          <span class="cost-row-desc">
            {prodUnits} units × {prodHoursPerDay} hrs × {DAYS_PER_MONTH} days × ${RATE_PER_UNIT_HR.toFixed(
              2,
            )}
          </span>
        </div>
        <span class="cost-row-amount">{fmtUSD(prodCost)}</span>
      </div>

      <!-- Development breakdown -->
      <div class="cost-row">
        <div class="cost-row-info">
          <span class="cost-row-label">Development</span>
          <span class="cost-row-desc">
            {devUnits} units × {devHoursPerDay} hrs × {DAYS_PER_MONTH} days × ${RATE_PER_UNIT_HR.toFixed(
              2,
            )}
          </span>
        </div>
        <span class="cost-row-amount">{fmtUSD(devCost)}</span>
      </div>

      <!-- Storage breakdown -->
      <div class="cost-row">
        <div class="cost-row-info">
          <span class="cost-row-label">Storage (GB)</span>
          <span class="cost-row-desc">
            {storageGB} GB − {FREE_STORAGE_GB} GB free = {billableStorageGB} GB ×
            ${STORAGE_RATE_PER_GB}
          </span>
        </div>
        <span class="cost-row-amount">{fmtUSD(storageCost)}</span>
      </div>

      <div class="cost-divider"></div>

      <!-- Monthly total -->
      <div class="cost-row">
        <span class="cost-row-label font-semibold">Monthly cost</span>
        <span class="cost-row-amount font-semibold">{fmtUSD(monthlyCost)}</span>
      </div>

      <!-- Available credit (trial/team plans only) -->
      {#if !isPro && availableCredit > 0}
        <div class="cost-row">
          <div class="cost-row-info">
            <span class="credit-label">Available credit</span>
            <span class="cost-row-desc">Applied to your first bill</span>
          </div>
          <span class="credit-amount">-{fmtUSD(availableCredit)}</span>
        </div>

        <div class="first-bill">
          <span class="first-bill-label">Estimated first bill</span>
          <span class="first-bill-amount">{fmtUSD(firstBill)}</span>
        </div>

        <span class="recurring-note">
          Then {fmtUSD(monthlyCost)}/mo based on current usage
        </span>
      {/if}

      <!-- Actions -->
      {#if !isPro}
        <button class="subscribe-btn" onclick={handleSubscribe}>
          Subscribe to Pro
        </button>
      {/if}

      <button class="contact-link" onclick={handleContactSales}>
        Contact sales for volume discounts
      </button>
    </div>
  </div>
</div>

<style lang="postcss">
  .estimate-page {
    @apply flex flex-col w-full max-w-5xl;
  }

  .back-link {
    @apply flex items-center gap-1 text-sm text-fg-secondary no-underline mb-1;
  }
  .back-link:hover {
    @apply text-fg-primary;
  }

  .page-title {
    @apply text-2xl font-semibold text-fg-primary mb-2;
  }

  .page-header {
    @apply flex items-center justify-between mb-6;
  }

  .prefill-note {
    @apply text-sm text-fg-tertiary;
  }

  .reset-btn {
    @apply flex items-center gap-1.5 text-sm font-medium text-primary-500 bg-transparent border-none cursor-pointer p-0;
  }
  .reset-btn:hover:not(:disabled) {
    @apply text-primary-600;
  }
  .reset-btn:disabled {
    @apply text-fg-disabled cursor-default;
  }

  .estimate-grid {
    @apply grid gap-6;
    grid-template-columns: 1fr 1fr;
    align-items: start;
  }

  /* Left panel */
  .input-panel {
    @apply border border-border rounded-xl bg-white p-6 flex flex-col;
  }

  .panel-header {
    @apply flex items-center justify-between mb-4;
  }

  .panel-title {
    @apply text-sm font-semibold text-fg-primary;
  }

  .unit-hint {
    @apply text-xs text-fg-tertiary;
  }

  .input-row {
    @apply flex items-center justify-between py-4;
  }

  .input-info {
    @apply flex flex-col gap-0.5;
  }

  .input-label {
    @apply text-sm font-semibold text-fg-primary;
  }

  .input-desc {
    @apply text-xs text-fg-tertiary;
  }

  .input-divider {
    @apply border-t border-border;
  }

  .input-row-sub {
    @apply py-2;
  }

  .input-label-sub {
    @apply text-xs text-fg-secondary;
  }

  /* Stepper */
  .stepper {
    @apply flex items-center border border-border rounded-md overflow-hidden shrink-0;
  }

  .stepper-btn {
    @apply w-8 h-8 flex items-center justify-center text-fg-secondary bg-transparent border-none cursor-pointer text-base;
  }
  .stepper-btn:hover:not(:disabled) {
    @apply bg-surface-subtle;
  }
  .stepper-btn:disabled {
    @apply text-fg-disabled cursor-default;
  }

  .stepper-input {
    @apply w-10 h-8 text-sm font-medium text-fg-primary border-x border-border tabular-nums text-center bg-transparent outline-none;
    -moz-appearance: textfield;
  }
  .stepper-input::-webkit-outer-spin-button,
  .stepper-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  /* Right panel */
  .cost-panel {
    @apply border border-border rounded-xl bg-white p-6 flex flex-col;
  }

  .cost-title {
    @apply text-sm font-semibold text-fg-primary mb-1;
  }

  .cost-total {
    @apply text-4xl font-bold text-fg-primary tabular-nums tracking-tight;
  }

  .cost-daily {
    @apply text-sm text-fg-tertiary mt-1 mb-4;
  }

  .cost-divider {
    @apply border-t border-border my-3;
  }

  .cost-row {
    @apply flex items-start justify-between py-2;
  }

  .cost-row-info {
    @apply flex flex-col gap-0.5;
  }

  .cost-row-label {
    @apply text-sm text-fg-primary;
  }

  .cost-row-desc {
    @apply text-xs text-fg-tertiary;
  }

  .cost-row-amount {
    @apply text-sm text-fg-primary tabular-nums text-right;
  }

  .credit-label {
    @apply text-sm font-medium;
    color: #16a34a;
  }

  .credit-amount {
    @apply text-sm font-medium tabular-nums text-right;
    color: #16a34a;
  }

  .first-bill {
    @apply flex items-center justify-between rounded-lg px-4 py-3 mt-2;
    background: var(--rill-colors-theme-primary-50, #ecf0ff);
  }

  .first-bill-label {
    @apply text-sm font-medium;
    color: #6366f1;
  }

  .first-bill-amount {
    @apply text-xl font-bold tabular-nums;
    color: #6366f1;
  }

  .recurring-note {
    @apply text-xs text-fg-tertiary mt-2;
  }

  .subscribe-btn {
    @apply w-full py-2.5 px-4 text-sm font-medium text-white rounded-none cursor-pointer mt-6 border-none;
    background: #6366f1;
  }
  .subscribe-btn:hover {
    background: #4f46e5;
  }

  .contact-link {
    @apply text-sm text-primary-500 bg-transparent border-none cursor-pointer mt-3 p-0;
  }
  .contact-link:hover {
    @apply text-primary-600 underline;
  }
</style>
