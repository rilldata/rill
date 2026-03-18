<script lang="ts">
  import {
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import {
    LIVE_CONNECT_TIERS,
    MANAGED_SLOT_TIERS,
    RILL_SLOT_TIERS,
    POPULAR_SLOTS,
    ALL_SLOTS,
    MANAGED_SLOT_RATE_PER_HR,
    CLUSTER_SLOT_RATE_PER_HR,
    RILL_SLOT_RATE_PER_HR,
    HOURS_PER_MONTH,
    STORAGE_RATE_PER_GB_PER_MONTH,
    INCLUDED_STORAGE_GB,
  } from "./slots-utils";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let currentSlots: number;
  export let isRillManaged: boolean;
  // Auto-detected slot count from the OLAP cluster (SQL-based detection).
  export let detectedSlots: number | undefined = undefined;
  // When true, the user can only view the detected tier and apply it (no selection).
  // required mode is no longer triggered (kept for template compatibility)
  export let required = false;
  export let viewOnly = false;
  // When true, uses new PRD v10 pricing (Free/Growth plans). False for legacy Team plans.
  export let useNewPricing = false;
  // Cluster Slots (auto-calculated from default OLAP connector; read-only). Only for Live Connect + new pricing.
  export let clusterSlots = 0;
  // Current Rill Slots (user-controlled). Only for Live Connect + new pricing.
  export let currentRillSlots = 0;
  // (infraSlots prop removed — no longer shown in the status UI)

  const POPULAR_RILL_MANAGED = POPULAR_SLOTS.map((s) => ({ slots: s }));
  const ALL_RILL_MANAGED = ALL_SLOTS.map((s) => ({ slots: s }));

  // All plans use the $0.15/slot/hr rate
  $: managedRate = MANAGED_SLOT_RATE_PER_HR;

  // For new pricing Live Connect: Rill Slots selection
  let selectedRillSlots = currentRillSlots;
  $: if (open && useNewPricing && !isRillManaged) {
    selectedRillSlots = currentRillSlots;
  }
  $: rillSlotsHasChanged = selectedRillSlots !== currentRillSlots;

  // Auto-detect matching tier from cluster memory
  $: detectedTierSlots = isRillManaged ? undefined : detectedSlots;

  // Rill-managed and self-managed: no minimum floor
  // When detectedTierSlots is set, the user cannot go below that tier.
  $: minimumSlots = detectedTierSlots ?? 0;

  // In required mode, pre-select detected tier or minimum; otherwise default to current
  $: minimumTierSlots = isRillManaged
    ? POPULAR_RILL_MANAGED[0].slots
    : LIVE_CONNECT_TIERS[0].slots;

  let selectedSlots = currentSlots;
  $: if (open) {
    if (viewOnly) {
      selectedSlots = detectedTierSlots ?? minimumTierSlots;
    } else if (required && currentSlots === 0) {
      selectedSlots = detectedTierSlots ?? minimumTierSlots;
    } else {
      selectedSlots = currentSlots;
    }
  }

  const updateProject = createAdminServiceUpdateProject();

  // GB usage for Rill-managed projects
  $: usageMetrics = isRillManaged
    ? getOrganizationUsageMetrics(organization)
    : undefined;
  $: projectUsageBytes =
    $usageMetrics?.data?.find((m) => m.project_name === project)?.size ?? 0;

  let showAllSizes = false;

  // Ensure the detected and current tiers always appear in the popular list
  $: popularSlotsWithExtras = (() => {
    let slots = [...POPULAR_SLOTS];
    const extras = [detectedTierSlots, currentSlots].filter(
      (s): s is number => s != null && s > 0,
    );
    for (const s of extras) {
      if (!slots.includes(s) && ALL_SLOTS.includes(s)) {
        slots.push(s);
      }
    }
    return slots.sort((a, b) => a - b);
  })();

  $: visibleRillManaged = showAllSizes
    ? ALL_RILL_MANAGED
    : ALL_RILL_MANAGED.filter((t) => popularSlotsWithExtras.includes(t.slots));

  $: visibleLiveConnect = showAllSizes
    ? LIVE_CONNECT_TIERS
    : LIVE_CONNECT_TIERS.filter((t) =>
        popularSlotsWithExtras.includes(t.slots),
      );

  $: hasChanged =
    useNewPricing && !isRillManaged
      ? rillSlotsHasChanged
      : selectedSlots !== currentSlots;

  async function applySlotChange() {
    try {
      // For new pricing Live Connect: prod_slots = cluster_slots + rill_slots.
      const newSlots =
        useNewPricing && !isRillManaged
          ? selectedRillSlots + clusterSlots
          : selectedSlots;
      await $updateProject.mutateAsync({
        org: organization,
        project,
        data: { prodSlots: String(newSlots) },
      });
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });
      eventBus.emit("notification", {
        message:
          useNewPricing && !isRillManaged
            ? `Rill Slots updated to ${selectedRillSlots}`
            : `Slots updated to ${newSlots}`,
      });
      open = false;
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message: axiosError.response?.data?.message ?? "Failed to update slots",
        type: "error",
      });
    }
  }
</script>

<Dialog.Root
  bind:open
  onOpenChange={(isOpen) => {
    if (required && !isOpen) return; // prevent closing when required
    open = isOpen;
  }}
  closeOnEscape={!required}
  closeOnOutsideClick={!required}
>
  <Dialog.Content class="max-w-2xl" noClose={required}>
    <Dialog.Header>
      <Dialog.Title>Manage Slots</Dialog.Title>
      <Dialog.Description>
        {#if viewOnly}
          Based on your OLAP cluster, we recommend the following slot
          configuration. <a
            href="/{organization}/-/settings/billing"
            class="text-primary-500 hover:underline">Start a Team plan</a
          > to customize your slot allocation.
        {:else if isRillManaged}
          Rill-managed projects are billed at ${managedRate}/slot/hr.{#if useNewPricing} Storage is ${STORAGE_RATE_PER_GB_PER_MONTH}/GB/month above {INCLUDED_STORAGE_GB}GB included.{:else} Data
          storage is charged separately based on usage.{/if} Monthly estimates assume
          ~{HOURS_PER_MONTH} hours/month.
        {:else if useNewPricing}
          Cluster Slots are auto-calculated from your OLAP cluster at ${CLUSTER_SLOT_RATE_PER_HR}/slot/hr.
          Add Rill Slots at ${RILL_SLOT_RATE_PER_HR}/slot/hr for extra performance or dev environments.
        {:else}
          Choose the slot tier that matches your OLAP cluster's resources. We
          auto-detect the minimum tier from your cluster configuration. You can
          increase slots if needed but cannot go below the detected minimum.
        {/if}
      </Dialog.Description>
    </Dialog.Header>

    {#if isRillManaged && projectUsageBytes > 0}
      <div class="usage-row">
        <span class="usage-label">Data usage</span>
        <span class="usage-value">{formatMemorySize(projectUsageBytes)}</span>
      </div>
    {/if}

    {#if isRillManaged}
      <!-- Rill-managed tier table -->
      <div class="tier-table">
        <div class="tier-header">
          <span class="tier-cell">Slots</span>
          <span class="tier-cell">$/slot/hr</span>
          <span class="tier-cell">Est. $/mo</span>
        </div>
        <div class="tier-list">
          {#each visibleRillManaged as tier}
            <button
              class="tier-row"
              class:tier-active={tier.slots === currentSlots ||
                (viewOnly && tier.slots === selectedSlots)}
              class:tier-selected={!viewOnly &&
                tier.slots === selectedSlots &&
                tier.slots !== currentSlots}
              class:tier-disabled={viewOnly && tier.slots !== selectedSlots}
              disabled={viewOnly && tier.slots !== selectedSlots}
              on:click={() => {
                if (!viewOnly) selectedSlots = tier.slots;
              }}
            >
              <span class="tier-cell">
                {tier.slots}
                {#if tier.slots === currentSlots}
                  <span class="current-badge">current</span>
                {/if}
              </span>
              <span class="tier-cell">${managedRate.toFixed(2)}</span>
              <span class="tier-cell">
                ~${Math.round(
                  tier.slots * managedRate * HOURS_PER_MONTH,
                ).toLocaleString()}
              </span>
            </button>
          {/each}
        </div>
      </div>
      {#if !viewOnly}
        <button
          class="show-all-btn"
          on:click={() => (showAllSizes = !showAllSizes)}
        >
          {showAllSizes ? "Show popular sizes" : "Show all sizes"}
        </button>
      {/if}
    {:else if useNewPricing}
      <!-- New pricing: Cluster Slots (read-only) + Rill Slots (user-controlled) -->
      <div class="dual-slot-section">
        <div class="slot-group">
          <div class="slot-group-header">
            <span class="slot-group-title">Cluster Slots</span>
            <span class="slot-group-subtitle">Auto-calculated from your OLAP cluster · read-only</span>
          </div>
          <div class="cluster-slot-display">
            <span class="cluster-slot-count">{clusterSlots}</span>
            <span class="cluster-slot-rate">
              @ ${CLUSTER_SLOT_RATE_PER_HR}/slot/hr
              (~${Math.round(clusterSlots * CLUSTER_SLOT_RATE_PER_HR * HOURS_PER_MONTH).toLocaleString()}/mo)
            </span>
          </div>
        </div>

        <div class="slot-group">
          <div class="slot-group-header">
            <span class="slot-group-title">Rill Slots</span>
            <span class="slot-group-subtitle">Additional slots for performance and dev environments</span>
          </div>
          <div class="tier-table">
            <div class="tier-header">
              <span class="tier-cell">Rill Slots</span>
              <span class="tier-cell">$/slot/hr</span>
              <span class="tier-cell">Est. $/mo</span>
            </div>
            <div class="tier-list">
              <!-- Option for 0 Rill Slots -->
              <button
                class="tier-row"
                class:tier-active={0 === currentRillSlots}
                class:tier-selected={0 === selectedRillSlots && 0 !== currentRillSlots}
                on:click={() => (selectedRillSlots = 0)}
              >
                <span class="tier-cell">
                  0
                  {#if 0 === currentRillSlots}
                    <span class="current-badge">current</span>
                  {/if}
                </span>
                <span class="tier-cell">-</span>
                <span class="tier-cell">$0</span>
              </button>
              {#each RILL_SLOT_TIERS.filter((t) => showAllSizes || POPULAR_SLOTS.includes(t.slots)) as tier}
                <button
                  class="tier-row"
                  class:tier-active={tier.slots === currentRillSlots}
                  class:tier-selected={tier.slots === selectedRillSlots && tier.slots !== currentRillSlots}
                  on:click={() => (selectedRillSlots = tier.slots)}
                >
                  <span class="tier-cell">
                    {tier.slots}
                    {#if tier.slots === currentRillSlots}
                      <span class="current-badge">current</span>
                    {/if}
                  </span>
                  <span class="tier-cell">${RILL_SLOT_RATE_PER_HR.toFixed(2)}</span>
                  <span class="tier-cell">~${tier.rillBill.toLocaleString()}</span>
                </button>
              {/each}
            </div>
          </div>
          <button
            class="show-all-btn"
            on:click={() => (showAllSizes = !showAllSizes)}
          >
            {showAllSizes ? "Show popular sizes" : "Show all sizes"}
          </button>
        </div>
      </div>

      <div class="total-row">
        <span class="total-label">Estimated total</span>
        <span class="total-value">
          ~${Math.round(
            (clusterSlots * CLUSTER_SLOT_RATE_PER_HR + selectedRillSlots * RILL_SLOT_RATE_PER_HR) * HOURS_PER_MONTH,
          ).toLocaleString()}/mo
        </span>
      </div>
    {:else}
      <!-- Legacy Live Connect tier table -->
      <div class="tier-table">
        <div class="tier-header">
          <span class="tier-cell tier-cell-wide">Cluster Size</span>
          <span class="tier-cell">Rill Slots</span>
          <span class="tier-cell">Estimated Rill $/mo</span>
        </div>
        <div class="tier-list">
          {#each visibleLiveConnect as tier}
            <button
              class="tier-row"
              class:tier-active={tier.slots === currentSlots ||
                (viewOnly && tier.slots === selectedSlots)}
              class:tier-selected={!viewOnly &&
                tier.slots === selectedSlots &&
                tier.slots !== currentSlots}
              class:tier-disabled={tier.slots < minimumSlots ||
                (viewOnly && tier.slots !== selectedSlots)}
              disabled={tier.slots < minimumSlots ||
                (viewOnly && tier.slots !== selectedSlots)}
              on:click={() => {
                if (!viewOnly) selectedSlots = tier.slots;
              }}
            >
              <span class="tier-cell tier-cell-wide">
                {tier.instance}
                {#if detectedTierSlots === tier.slots}
                  <span class="detected-badge">detected</span>
                {/if}
              </span>
              <span class="tier-cell">
                {tier.slots}
                {#if tier.slots === currentSlots}
                  <span class="current-badge">current</span>
                {/if}
              </span>
              <span class="tier-cell">~${tier.rillBill.toLocaleString()}</span>
            </button>
          {/each}
        </div>
      </div>
      {#if !viewOnly}
        <button
          class="show-all-btn"
          on:click={() => (showAllSizes = !showAllSizes)}
        >
          {showAllSizes ? "Show popular sizes" : "Show all sizes"}
        </button>
      {/if}
      <p class="chc-note">
        Estimated costs are calculated at a full month. Billing is charged at
        compute/hr, therefore variable based on your needs. Select the tier that best
        matches your cluster's memory and vCPU allocation.
      </p>
    {/if}

    <div class="footer">
      {#if !required}
        <button class="cancel-btn" on:click={() => (open = false)}>
          Cancel
        </button>
      {/if}
      <button
        class="apply-btn"
        disabled={(viewOnly
          ? false
          : required
            ? selectedSlots === 0
            : !hasChanged) || $updateProject.isPending}
        on:click={applySlotChange}
      >
        {#if $updateProject.isPending}
          Updating...
        {:else}
          Apply
        {/if}
      </button>
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .usage-row {
    @apply flex items-center gap-2 mb-3;
  }
  .usage-label {
    @apply text-sm text-fg-secondary w-28 shrink-0;
  }
  .usage-value {
    @apply text-sm text-fg-primary font-medium;
  }
  .tier-table {
    @apply border border-border rounded-md overflow-hidden;
  }
  .tier-list {
    @apply max-h-[280px] overflow-y-auto;
  }
  .tier-header {
    @apply flex bg-surface-subtle text-xs font-semibold text-fg-secondary uppercase tracking-wide;
  }
  .tier-header .tier-cell {
    @apply px-3 py-2;
  }
  .tier-row {
    @apply flex text-sm border-t border-border w-full text-left bg-transparent cursor-pointer;
  }
  .tier-row:hover:not(:disabled):not(.tier-active):not(.tier-selected) {
    @apply bg-surface-subtle;
  }
  .tier-row .tier-cell {
    @apply px-3 py-2;
  }
  .tier-active {
    @apply bg-primary-50;
  }
  .tier-selected {
    @apply bg-primary-100;
  }
  .tier-disabled {
    @apply opacity-40 cursor-not-allowed;
  }
  .tier-cell {
    @apply flex-1 flex items-center gap-1.5;
  }
  .tier-cell-wide {
    @apply flex-[2];
  }
  .current-badge {
    @apply text-[10px] text-primary-600 bg-primary-100 px-1.5 py-0.5 rounded-full leading-none font-medium;
  }
  .detected-badge {
    @apply text-[10px] text-green-700 bg-green-100 px-1.5 py-0.5 rounded-full leading-none font-medium;
  }
  .show-all-btn {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0 mt-2;
  }
  .show-all-btn:hover {
    @apply text-primary-600;
  }
  .chc-note {
    @apply text-xs text-fg-tertiary mt-2 italic;
  }
  .footer {
    @apply flex justify-end gap-2 mt-4;
  }
  .cancel-btn {
    @apply text-sm text-fg-secondary bg-transparent border border-border rounded-md px-3 py-1.5 cursor-pointer;
  }
  .cancel-btn:hover {
    @apply bg-surface-subtle;
  }
  .apply-btn {
    @apply text-sm text-white bg-primary-500 border-none rounded-md px-3 py-1.5 cursor-pointer font-medium;
  }
  .apply-btn:hover {
    @apply bg-primary-600;
  }
  .apply-btn:disabled {
    @apply opacity-50 cursor-not-allowed;
  }
  .dual-slot-section {
    @apply flex flex-col gap-5;
  }
  .slot-group {
    @apply flex flex-col gap-2;
  }
  .slot-group-header {
    @apply flex flex-col gap-0.5;
  }
  .slot-group-title {
    @apply text-sm font-semibold text-fg-primary;
  }
  .slot-group-subtitle {
    @apply text-xs text-fg-tertiary;
  }
  .cluster-slot-display {
    @apply flex items-center gap-3 px-3 py-2.5 bg-surface-subtle rounded-md border border-border;
  }
  .cluster-slot-count {
    @apply text-lg font-semibold text-fg-primary;
  }
  .cluster-slot-rate {
    @apply text-sm text-fg-secondary;
  }
  .total-row {
    @apply flex items-center justify-between px-3 py-2.5 mt-3 bg-surface-subtle rounded-md border border-border;
  }
  .total-label {
    @apply text-sm font-semibold text-fg-primary;
  }
  .total-value {
    @apply text-sm font-semibold text-fg-primary;
  }
</style>
