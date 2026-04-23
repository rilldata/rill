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
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ArrowLeft, RotateCw, ChevronDown } from "lucide-svelte";
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
  let storageAmount = $state(1);
  let storageUnit = $state<"GB" | "TB">("GB");

  // Billing is calculated in integer GB; TB converts at 1 TB = 1000 GB
  // (decimal, SI). Floor defends against transient fractional input values.
  let storageGB = $derived(
    Math.max(
      Math.floor(storageUnit === "TB" ? storageAmount * 1000 : storageAmount),
      0,
    ),
  );

  // Pre-fill from current deployment once data loads
  let prefilled = $state(false);
  $effect(() => {
    if (!prefilled && !$projectsQuery.isLoading && projects.length > 0) {
      prodUnits = Math.max(currentProdSlots, 2);
      devUnits = currentDevSlots;
      storageAmount = Math.max(currentStorageGB, 1);
      storageUnit = "GB";
      prefilled = true;
    }
  });

  function resetToUsage() {
    prodUnits = Math.max(currentProdSlots, 2);
    prodHoursPerDay = 24;
    devUnits = currentDevSlots;
    devHoursPerDay = 24;
    storageAmount = Math.max(currentStorageGB, 1);
    storageUnit = "GB";
  }

  // Pricing constants
  const RATE_PER_UNIT_HR = 0.15;
  const STORAGE_RATE_PER_GB = 1;
  const FREE_STORAGE_GB = 1;
  const DAYS_PER_MONTH = 30;
  const DEFAULT_ACTIVE_HOURS = 8;

  // "Always on" mirrors hours === 24; toggling restores a sensible default.
  let prodAlwaysOn = $derived(prodHoursPerDay === 24);
  let devAlwaysOn = $derived(devHoursPerDay === 24);

  function toggleProdAlwaysOn() {
    prodHoursPerDay = prodAlwaysOn ? DEFAULT_ACTIVE_HOURS : 24;
  }
  function toggleDevAlwaysOn() {
    devHoursPerDay = devAlwaysOn ? DEFAULT_ACTIVE_HOURS : 24;
  }

  // Hours used by the estimate; clamped so mid-typing overflows (e.g. "25")
  // don't produce out-of-range costs or breakdown text.
  let effectiveProdHours = $derived(Math.min(Math.max(prodHoursPerDay, 0), 24));
  let effectiveDevHours = $derived(Math.min(Math.max(devHoursPerDay, 0), 24));

  // Cost calculations
  let prodCost = $derived(
    prodUnits * effectiveProdHours * DAYS_PER_MONTH * RATE_PER_UNIT_HR,
  );
  let devCost = $derived(
    devUnits * effectiveDevHours * DAYS_PER_MONTH * RATE_PER_UNIT_HR,
  );
  let billableStorageGB = $derived(Math.max(storageGB - FREE_STORAGE_GB, 0));
  let storageCost = $derived(billableStorageGB * STORAGE_RATE_PER_GB);
  let monthlyCost = $derived(prodCost + devCost + storageCost);
  let dailyCost = $derived(monthlyCost / DAYS_PER_MONTH);

  // TODO: Wire to billing API when available
  let availableCredit = $derived(100);
  let youPay = $derived(Math.max(monthlyCost - availableCredit, 0));

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
      storageAmount === Math.max(currentStorageGB, 1) &&
      storageUnit === "GB",
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
  <!-- Back link -->
  <a href="/{organization}/-/settings/billing" class="back-link">
    <ArrowLeft class="w-4 h-4" />
    Billing
  </a>
  <h1 class="page-title">Estimate your cost</h1>

  <p class="pricing-hint">
    Pricing: <strong>$0.15</strong> per unit/hr · <strong>$1</strong> per GB/month.
    1 unit = 4 GiB RAM, 1 vCPU.
  </p>

  <div class="estimate-grid">
    <!-- LEFT: Compute unit and storage -->
    <div class="input-panel">
      <div class="panel-header">
        <div class="panel-header-text">
          <span class="panel-title">Compute unit and storage</span>
          <span class="panel-sub">Pre-filled from your current deployment.</span>
        </div>
        <button
          class="reset-btn"
          onclick={resetToUsage}
          disabled={isCurrentConfig}
          type="button"
        >
          <RotateCw class="w-4 h-4" />
          Reset
        </button>
      </div>

      <!-- Production row -->
      <div class="field-row">
        <div class="field-col">
          <span class="field-label">Production unit</span>
          <div class="stepper">
            <button
              class="stepper-btn"
              onclick={() => prodUnits--}
              disabled={prodUnits <= 2}
              aria-label="Decrease production units"
            >
              −
            </button>
            <input
              class="stepper-input"
              type="number"
              bind:value={prodUnits}
              min="2"
              onblur={(e) => clampInput(e, 2, 9999, (v) => (prodUnits = v))}
              aria-label="Production units"
            />
            <button
              class="stepper-btn"
              onclick={() => prodUnits++}
              aria-label="Increase production units"
            >
              +
            </button>
          </div>
          <span class="field-hint">Minimum 2 units</span>
        </div>

        <div class="field-col">
          <span class="field-label">Active hours per day</span>
          <div class="stepper">
            <button
              class="stepper-btn"
              onclick={() => prodHoursPerDay--}
              disabled={prodHoursPerDay <= 1 || prodAlwaysOn}
              aria-label="Decrease production active hours"
            >
              −
            </button>
            <input
              class="stepper-input"
              type="number"
              bind:value={prodHoursPerDay}
              min="1"
              max="24"
              disabled={prodAlwaysOn}
              onblur={(e) => clampInput(e, 1, 24, (v) => (prodHoursPerDay = v))}
              aria-label="Production active hours"
            />
            <button
              class="stepper-btn"
              onclick={() => prodHoursPerDay++}
              disabled={prodHoursPerDay >= 24 || prodAlwaysOn}
              aria-label="Increase production active hours"
            >
              +
            </button>
          </div>
          <label class="checkbox-label">
            <input
              type="checkbox"
              class="checkbox"
              checked={prodAlwaysOn}
              onchange={toggleProdAlwaysOn}
            />
            Always on
          </label>
        </div>
      </div>

      <div class="field-divider"></div>

      <!-- Development row -->
      <div class="field-row">
        <div class="field-col">
          <span class="field-label">Development unit</span>
          <div class="stepper">
            <button
              class="stepper-btn"
              onclick={() => devUnits--}
              disabled={devUnits <= 0}
              aria-label="Decrease development units"
            >
              −
            </button>
            <input
              class="stepper-input"
              type="number"
              bind:value={devUnits}
              min="0"
              onblur={(e) => clampInput(e, 0, 9999, (v) => (devUnits = v))}
              aria-label="Development units"
            />
            <button
              class="stepper-btn"
              onclick={() => devUnits++}
              aria-label="Increase development units"
            >
              +
            </button>
          </div>
        </div>

        <div class="field-col">
          <span class="field-label">
            Active hours per day
            <Tooltip location="top" alignment="middle" distance={8}>
              <span class="info-icon"><InfoCircle size="14px" /></span>
              <TooltipContent maxWidth="240px" slot="tooltip-content">
                Deployments hibernate when inactive, so you're only billed for
                active hours.
              </TooltipContent>
            </Tooltip>
          </span>
          <div class="stepper">
            <button
              class="stepper-btn"
              onclick={() => devHoursPerDay--}
              disabled={devHoursPerDay <= 1 || devAlwaysOn}
              aria-label="Decrease development active hours"
            >
              −
            </button>
            <input
              class="stepper-input"
              type="number"
              bind:value={devHoursPerDay}
              min="1"
              max="24"
              disabled={devAlwaysOn}
              onblur={(e) => clampInput(e, 1, 24, (v) => (devHoursPerDay = v))}
              aria-label="Development active hours"
            />
            <button
              class="stepper-btn"
              onclick={() => devHoursPerDay++}
              disabled={devHoursPerDay >= 24 || devAlwaysOn}
              aria-label="Increase development active hours"
            >
              +
            </button>
          </div>
          <label class="checkbox-label">
            <input
              type="checkbox"
              class="checkbox"
              checked={devAlwaysOn}
              onchange={toggleDevAlwaysOn}
            />
            Always on
          </label>
        </div>
      </div>

      <div class="field-divider"></div>

      <!-- Storage row -->
      <div class="storage-row">
        <div class="storage-info">
          <span class="panel-title">Storage</span>
          <span class="panel-sub">1 GB included free</span>
        </div>
        <div class="storage-controls">
          <input
            class="storage-input"
            type="number"
            bind:value={storageAmount}
            min="1"
            step="1"
            onblur={(e) => clampInput(e, 1, 999999, (v) => (storageAmount = v))}
            onkeydown={(e) => {
              if (e.key === "." || e.key === "e" || e.key === "-") {
                e.preventDefault();
              }
            }}
            aria-label="Storage amount"
          />
          <DropdownMenu.Root>
            <DropdownMenu.Trigger class="storage-unit-trigger">
              {storageUnit}
              <ChevronDown class="w-4 h-4" />
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end" class="min-w-[64px]">
              <DropdownMenu.Item onclick={() => (storageUnit = "GB")}>
                GB
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => (storageUnit = "TB")}>
                TB
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      </div>
    </div>

    <!-- RIGHT: Cost summary -->
    <div class="cost-panel">
      <div class="cost-header">
        <span class="cost-title">Estimate monthly cost</span>
        <span class="cost-total">{fmtUSD(monthlyCost)}</span>
        <span class="cost-daily">~{fmtUSD(dailyCost)}/day</span>
      </div>

      <div class="cost-breakdown">
        <div class="cost-row">
          <div class="cost-row-info">
            <span class="cost-row-label">Production</span>
            <span class="cost-row-desc">
              <strong>{prodUnits}</strong> units × <strong
                >{effectiveProdHours}</strong
              > hrs × <strong>{DAYS_PER_MONTH}</strong> days × <strong
                >${RATE_PER_UNIT_HR.toFixed(2)}</strong
              >
            </span>
          </div>
          <span class="cost-row-amount">{fmtUSD(prodCost)}</span>
        </div>

        <div class="cost-row">
          <div class="cost-row-info">
            <span class="cost-row-label">Development</span>
            <span class="cost-row-desc">
              <strong>{devUnits}</strong> units × <strong
                >{effectiveDevHours}</strong
              > hrs × <strong>{DAYS_PER_MONTH}</strong> days × <strong
                >${RATE_PER_UNIT_HR.toFixed(2)}</strong
              >
            </span>
          </div>
          <span class="cost-row-amount">{fmtUSD(devCost)}</span>
        </div>

        <div class="cost-row">
          <div class="cost-row-info">
            <span class="cost-row-label">Storage (GB)</span>
            <span class="cost-row-desc">
              <strong>{billableStorageGB}</strong> billable GB × <strong
                >${STORAGE_RATE_PER_GB}</strong
              >/GB
            </span>
          </div>
          <span class="cost-row-amount">{fmtUSD(storageCost)}</span>
        </div>
      </div>

      <div class="cost-footer">
        <div class="monthly-row">
          <span class="monthly-label">Monthly cost</span>
          <span class="monthly-amount">{fmtUSD(monthlyCost)}</span>
        </div>

        {#if !isPro && availableCredit > 0}
          <div class="credit-row">
            <div class="credit-info">
              <span class="credit-label">Available credit</span>
              <span class="credit-desc">Applied to your first bill</span>
            </div>
            <span class="credit-amount">-{fmtUSD(availableCredit)}</span>
          </div>

          <div class="you-pay">
            <span class="you-pay-label">You pay</span>
            <span class="you-pay-amount">{fmtUSD(youPay)}</span>
          </div>

          <span class="recurring-note">
            Then {fmtUSD(monthlyCost)}/mo at this configuration.
          </span>
        {/if}
      </div>

      <div class="cost-actions">
        {#if !isPro}
          <button class="subscribe-btn" onclick={handleSubscribe} type="button">
            Subscribe to Pro
          </button>
        {/if}
        <button
          class="contact-link"
          onclick={handleContactSales}
          type="button"
        >
          Contact sales for volume discounts
        </button>
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .estimate-page {
    @apply flex flex-col items-center w-full pt-6 pb-12 px-4 gap-3;
  }

  .back-link {
    @apply flex items-center gap-1 text-base font-medium text-fg-secondary no-underline w-[900px] pt-3;
  }
  .back-link:hover {
    @apply text-fg-primary;
  }

  .page-title {
    @apply text-2xl font-semibold text-fg-secondary w-[900px];
  }

  .pricing-hint {
    @apply text-sm text-fg-secondary w-[900px] mt-2;
  }
  .pricing-hint strong {
    @apply font-semibold;
  }

  .estimate-grid {
    @apply flex items-start gap-3 w-[900px];
  }

  /* Left panel */
  .input-panel {
    @apply flex-1 flex flex-col gap-8 bg-white border border-border rounded-lg p-6;
    box-shadow:
      0 1px 3px 0 rgba(0, 0, 0, 0.1),
      0 1px 2px 0 rgba(0, 0, 0, 0.1);
  }

  .panel-header {
    @apply flex items-start gap-2 w-full;
  }

  .panel-header-text {
    @apply flex flex-col gap-2;
  }

  .panel-title {
    @apply text-base font-semibold text-fg-primary leading-none;
  }

  .panel-sub {
    @apply text-sm text-fg-tertiary leading-none;
  }

  .reset-btn {
    @apply flex items-center gap-2 text-xs font-medium text-primary-500 bg-transparent border-none cursor-pointer ml-auto p-0;
  }
  .reset-btn:hover:not(:disabled) {
    @apply text-primary-600;
  }
  .reset-btn:disabled {
    @apply text-fg-disabled cursor-default;
  }

  /* Field rows (Production / Development) */
  .field-row {
    @apply flex gap-6 w-full pb-6 border-b border-border;
  }

  .field-row:last-of-type {
    @apply border-b-0 pb-0;
  }

  .field-col {
    @apply flex-1 flex flex-col gap-2;
  }

  .field-label {
    @apply inline-flex items-center gap-2 text-sm font-semibold text-fg-secondary leading-none;
  }

  .info-icon {
    @apply text-fg-tertiary flex;
  }

  .field-hint {
    @apply text-sm text-fg-tertiary leading-none;
  }

  .field-divider {
    @apply -my-4;
  }

  /* Stepper */
  .stepper {
    @apply flex items-center w-full;
  }

  .stepper-btn {
    @apply w-9 h-9 flex items-center justify-center text-base text-fg-primary bg-white border border-border cursor-pointer;
  }
  .stepper-btn:first-child {
    @apply rounded-l-md;
  }
  .stepper-btn:last-child {
    @apply rounded-r-md;
  }
  .stepper-btn:hover:not(:disabled) {
    @apply bg-surface-subtle;
  }
  .stepper-btn:disabled {
    @apply text-fg-disabled cursor-default;
  }

  .stepper-input {
    @apply flex-1 h-9 text-sm text-fg-primary text-center bg-white border-y border-border tabular-nums outline-none min-w-0;
    -moz-appearance: textfield;
  }
  .stepper-input::-webkit-outer-spin-button,
  .stepper-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  .stepper-input:disabled {
    @apply bg-surface-subtle text-fg-disabled;
  }

  /* Checkbox */
  .checkbox-label {
    @apply flex items-center gap-2 text-sm font-medium text-fg-primary cursor-pointer select-none;
  }

  .checkbox {
    @apply w-4 h-4 rounded border border-border bg-white cursor-pointer accent-primary-500;
  }

  /* Storage row */
  .storage-row {
    @apply flex items-start gap-3 w-full;
  }

  .storage-info {
    @apply flex-1 flex flex-col gap-1.5;
  }

  .storage-controls {
    @apply flex items-center gap-2;
  }

  .storage-input {
    @apply w-16 h-9 px-3 py-1 text-sm text-fg-primary bg-white border border-border rounded-sm tabular-nums outline-none;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    -moz-appearance: textfield;
  }
  .storage-input::-webkit-outer-spin-button,
  .storage-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  :global(.storage-unit-trigger) {
    @apply flex items-center gap-2 h-9 px-4 text-sm font-medium text-fg-primary bg-white border border-border rounded-sm cursor-pointer outline-none;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  }
  :global(.storage-unit-trigger:hover) {
    @apply bg-surface-subtle;
  }
  :global(.storage-unit-trigger:focus-visible) {
    @apply ring-2 ring-primary-400 ring-offset-2;
  }

  /* Right panel */
  .cost-panel {
    @apply w-96 flex flex-col gap-6 bg-white border border-border rounded-lg p-6;
    box-shadow:
      0 1px 3px 0 rgba(0, 0, 0, 0.1),
      0 1px 2px -1px rgba(0, 0, 0, 0.1);
  }

  .cost-header {
    @apply flex flex-col gap-2 pb-6 border-b border-border;
  }

  .cost-title {
    @apply text-sm font-semibold text-fg-primary leading-none;
  }

  .cost-total {
    @apply text-4xl font-semibold text-fg-primary tabular-nums leading-[44px];
  }

  .cost-daily {
    @apply text-xs font-medium text-fg-tertiary leading-none;
  }

  .cost-breakdown {
    @apply flex flex-col gap-3;
  }

  .cost-row {
    @apply flex items-start gap-6 w-full;
  }

  .cost-row-info {
    @apply flex-1 flex flex-col gap-1.5 min-w-0;
  }

  .cost-row-label {
    @apply text-xs font-semibold text-fg-primary leading-none;
  }

  .cost-row-desc {
    @apply text-xs text-fg-tertiary leading-4;
  }
  .cost-row-desc strong {
    @apply font-semibold;
  }

  .cost-row-amount {
    @apply text-xs font-semibold text-fg-tertiary tabular-nums whitespace-nowrap;
  }

  .cost-footer {
    @apply flex flex-col gap-3;
  }

  .monthly-row {
    @apply flex items-center gap-6 py-4 border-y border-border;
  }

  .monthly-label {
    @apply flex-1 text-xs font-semibold text-fg-primary leading-none;
  }

  .monthly-amount {
    @apply text-base font-semibold text-fg-secondary tabular-nums;
  }

  .credit-row {
    @apply flex items-start gap-6;
  }

  .credit-info {
    @apply flex-1 flex flex-col gap-1.5 min-w-0;
  }

  .credit-label {
    @apply text-xs font-semibold text-green-700 leading-none;
  }

  .credit-desc {
    @apply text-xs text-fg-tertiary leading-4;
  }

  .credit-amount {
    @apply text-base font-semibold text-green-700 tabular-nums whitespace-nowrap;
  }

  .you-pay {
    @apply flex items-center gap-6 py-2 px-4 rounded-md bg-primary-50;
  }

  .you-pay-label {
    @apply flex-1 text-base font-semibold text-primary-500 leading-none;
  }

  .you-pay-amount {
    @apply text-2xl font-semibold text-primary-500 tabular-nums;
    line-height: 36px;
  }

  .recurring-note {
    @apply text-xs text-fg-tertiary text-center leading-4;
  }

  .cost-actions {
    @apply flex flex-col gap-1.5;
  }

  .subscribe-btn {
    @apply w-full h-9 px-4 text-sm font-medium text-white bg-primary-500 rounded-sm border-none cursor-pointer;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  }
  .subscribe-btn:hover {
    @apply bg-primary-600;
  }

  .contact-link {
    @apply w-full h-9 px-4 text-sm font-medium text-primary-500 bg-transparent border-none cursor-pointer;
  }
  .contact-link:hover {
    @apply text-primary-600 underline;
  }
</style>
