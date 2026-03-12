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

  // Live Connect tiers: Rill bill is ~20% of infrastructure cost
  const LIVE_CONNECT_TIERS = [
    { slots: 4, instance: "Basic (8GB / 2vCPU) × 2", rillBill: 88 },
    { slots: 6, instance: "Basic (12GB / 3vCPU) × 2", rillBill: 131 },
    { slots: 8, instance: "Scale (16GB / 4vCPU) × 2", rillBill: 175 },
    { slots: 16, instance: "Scale (32GB / 8vCPU) × 2", rillBill: 350 },
    { slots: 32, instance: "Scale (64GB / 16vCPU) × 2", rillBill: 701 },
    { slots: 60, instance: "Scale (120GB / 30vCPU) × 2", rillBill: 1314 },
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

  // For Live Connect: can scale up but not below initial cluster baseline
  $: minimumSlots = isRillManaged ? 0 : currentSlots;

  let selectedSlots = currentSlots;
  $: if (open) selectedSlots = currentSlots;

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

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Manage Slots</Dialog.Title>
      <Dialog.Description>
        {#if isRillManaged}
          Rill-managed projects are billed at ${SLOT_RATE_PER_HR}/slot/hr. Data
          storage is charged separately based on usage. Monthly estimates assume
          ~{HOURS_PER_MONTH} hours/month.
        {:else}
          We provide minimum slots based on your ClickHouse Cloud / Self Managed
          cluster settings. You're free to increase if required but cannot go
          below your cluster's minimum specs.
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
            class:tier-active={tier.slots === currentSlots}
            class:tier-selected={tier.slots === selectedSlots &&
              tier.slots !== currentSlots}
            on:click={() => (selectedSlots = tier.slots)}
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
          <span class="tier-cell tier-cell-wide">CHC Cluster</span>
          <span class="tier-cell">Rill Slots</span>
          <span class="tier-cell">Rill $/mo</span>
        </div>
        {#each LIVE_CONNECT_TIERS as tier}
          <button
            class="tier-row"
            class:tier-active={tier.slots === currentSlots}
            class:tier-selected={tier.slots === selectedSlots &&
              tier.slots !== currentSlots}
            class:tier-disabled={tier.slots < minimumSlots}
            disabled={tier.slots < minimumSlots}
            on:click={() => (selectedSlots = tier.slots)}
          >
            <span class="tier-cell tier-cell-wide">
              {tier.instance}
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
        Cluster specs shown for ClickHouse Cloud reference. Self-managed users
        can ignore cluster details.
      </p>
    {/if}

    <div class="footer">
      <button class="cancel-btn" on:click={() => (open = false)}>
        Cancel
      </button>
      <button
        class="apply-btn"
        disabled={!hasChanged || $updateProject.isPending}
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
