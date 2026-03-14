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

  export let open = false;
  export let organization: string;
  export let project: string;
  export let currentSlots: number;
  export let isRillManaged: boolean;
  // Whether the OLAP connector is ClickHouse Cloud (vs generic self-managed Live Connect).
  export let isClickHouseCloud = false;
  // When required, the user cannot dismiss the modal and must select + apply slots.
  export let required = false;
  // ClickHouse Cloud cluster memory (GB per replica) for auto-detecting the right tier.
  export let detectedMemoryGb: number | undefined = undefined;
  // When true, the user can only view the detected tier and apply it (no selection).
  export let viewOnly = false;

  // Live Connect tiers: Rill bill is ~20% of infrastructure cost
  const LIVE_CONNECT_TIERS = [
    { slots: 4, instance: "8GiB / 2vCPU", rillBill: 99 }, 
    { slots: 6, instance: "12GiB / 3vCPU", rillBill: 130 },
    { slots: 8, instance: "16GiB / 4vCPU", rillBill: 175 },
    { slots: 16, instance: "32GiB / 8vCPU", rillBill: 350 },
    { slots: 32, instance: "64GiB / 16vCPU", rillBill: 700 },
    { slots: 60, instance: "120GiB / 30vCPU", rillBill: 1300 },
  ];

  // Rill-managed tiers: billed per slot/hr; data charged separately
  const RILL_MANAGED_TIERS = [
    { slots: 4 },
    { slots: 6 },
    { slots: 8 },
    { slots: 16 },
    { slots: 32 },
    { slots: 60 },
  ];

  const SLOT_RATE_PER_HR = 0.03;
  const HOURS_PER_MONTH = 730; // ~365.25 * 24 / 12

  // Auto-detect matching tier from CHC cluster memory (memory per replica)
  $: detectedTierSlots = (() => {
    if (!detectedMemoryGb || isRillManaged) return undefined;
    // Find the tier whose per-replica memory (slots * 2 GB) is closest to detected
    const match = LIVE_CONNECT_TIERS.reduce((best, tier) => {
      const tierMemory = tier.slots * 2; // GB per replica
      const bestMemory = best.slots * 2;
      return Math.abs(tierMemory - detectedMemoryGb) <
        Math.abs(bestMemory - detectedMemoryGb)
        ? tier
        : best;
    });
    return match.slots;
  })();

  // Rill-managed and self-managed: no minimum floor
  // CHC (detectedTierSlots set): can downgrade below current but not below detected tier
  $: minimumSlots = detectedTierSlots ?? 0;

  // In required mode, pre-select detected tier or minimum; otherwise default to current
  $: minimumTierSlots = isRillManaged
    ? RILL_MANAGED_TIERS[0].slots
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

  $: hasChanged = selectedSlots !== currentSlots;

  async function applySlotChange() {
    try {
      await $updateProject.mutateAsync({
        org: organization,
        project,
        data: { prodSlots: String(selectedSlots) },
      });
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });
      eventBus.emit("notification", {
        message: `Slots updated to ${selectedSlots}`,
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
          {#if isClickHouseCloud}
            Based on your ClickHouse Cloud cluster, we recommend the following
            slot configuration. <a
              href="/{organization}/-/settings/billing"
              class="text-primary-500 hover:underline"
              >Start a Growth plan</a
            > to customize your slot allocation.
          {:else}
            Based on your OLAP cluster, we recommend the following
            slot configuration. <a
              href="/{organization}/-/settings/billing"
              class="text-primary-500 hover:underline"
              >Start a Growth plan</a
            > to customize your slot allocation.
          {/if}
        {:else if isRillManaged}
          Rill-managed projects are billed at ${SLOT_RATE_PER_HR}/slot/hr. Data
          storage is charged separately based on usage. Monthly estimates assume
          ~{HOURS_PER_MONTH} hours/month.
        {:else if isClickHouseCloud}
          Slots are matched to your ClickHouse Cloud cluster size. We
          auto-detect the minimum tier from your service configuration. You can
          increase slots if needed but cannot go below the detected minimum.
        {:else}
          Choose the slot tier that matches your OLAP cluster's resources. You
          can increase slots at any time to handle larger workloads.
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
        {#each RILL_MANAGED_TIERS as tier}
          <button
            class="tier-row"
            class:tier-active={tier.slots === currentSlots || (viewOnly && tier.slots === selectedSlots)}
            class:tier-selected={!viewOnly && tier.slots === selectedSlots &&
              tier.slots !== currentSlots}
            class:tier-disabled={viewOnly && tier.slots !== selectedSlots}
            disabled={viewOnly && tier.slots !== selectedSlots}
            on:click={() => { if (!viewOnly) selectedSlots = tier.slots; }}
          >
            <span class="tier-cell">
              {tier.slots}
              {#if tier.slots === currentSlots}
                <span class="current-badge">current</span>
              {/if}
            </span>
            <span class="tier-cell">${SLOT_RATE_PER_HR.toFixed(2)}</span>
            <span class="tier-cell">
              ~${Math.round(
                tier.slots * SLOT_RATE_PER_HR * HOURS_PER_MONTH,
              ).toLocaleString()}
            </span>
          </button>
        {/each}
      </div>
    {:else}
      <!-- Live Connect tier table -->
      <div class="tier-table">
        <div class="tier-header">
          <span class="tier-cell tier-cell-wide">
            {isClickHouseCloud ? "CHC Cluster" : "Cluster Size"}
          </span>
          <span class="tier-cell">Rill Slots</span>
          <span class="tier-cell">Estimated Rill $/mo</span>
        </div>
        {#each LIVE_CONNECT_TIERS as tier}
          <button
            class="tier-row"
            class:tier-active={tier.slots === currentSlots || (viewOnly && tier.slots === selectedSlots)}
            class:tier-selected={!viewOnly && tier.slots === selectedSlots &&
              tier.slots !== currentSlots}
            class:tier-disabled={tier.slots < minimumSlots || (viewOnly && tier.slots !== selectedSlots)}
            disabled={tier.slots < minimumSlots || (viewOnly && tier.slots !== selectedSlots)}
            on:click={() => { if (!viewOnly) selectedSlots = tier.slots; }}
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
      <p class="chc-note">
        Estimated costs are calculated at a full month. Billing is charged at
        compute/hr, therefore variable based on your needs.
      </p>
      {#if isClickHouseCloud}
        <p class="chc-note">
          Cluster specs are auto-detected from your ClickHouse Cloud service
          and assume 2 replicas.
        </p>
      {:else}
        <p class="chc-note">
          Select the tier that best matches your cluster's memory and vCPU
          allocation. The cluster size column is for reference only.
        </p>
      {/if}
    {/if}

    <div class="footer">
      {#if !required}
        <button class="cancel-btn" on:click={() => (open = false)}>
          Cancel
        </button>
      {/if}
      <button
        class="apply-btn"
        disabled={(viewOnly ? false : required ? selectedSlots === 0 : !hasChanged) || $updateProject.isPending}
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
</style>
