<script lang="ts">
  import {
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import {
    SLOT_TIERS,
    POPULAR_SLOTS,
    ALL_SLOTS,
    SLOT_RATE_PER_HR,
    HOURS_PER_MONTH,
    DEFAULT_MANAGED_SLOTS,
    DEFAULT_SELF_MANAGED_SLOTS,
  } from "./slots-utils";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let currentSlots: number;
  export let isRillManaged = true;
  export let viewOnly = false;

  // Minimum slots: 2 for Rill-managed, 4 for self-managed
  $: minSlots = isRillManaged
    ? DEFAULT_MANAGED_SLOTS
    : DEFAULT_SELF_MANAGED_SLOTS;

  let selectedSlots = currentSlots;
  $: if (open) {
    selectedSlots = currentSlots;
    showAllSizes = false;
  }

  let showAllSizes = false;

  const updateProject = createAdminServiceUpdateProject();

  // Filter tiers to only show slots >= minimum
  $: availableTiers = SLOT_TIERS.filter((t) => t.slots >= minSlots);

  // Ensure the current slot count always appears in the popular list
  $: popularSlotsWithExtras = (() => {
    let slots = POPULAR_SLOTS.filter((s) => s >= minSlots);
    if (
      currentSlots >= minSlots &&
      !slots.includes(currentSlots) &&
      ALL_SLOTS.includes(currentSlots)
    ) {
      slots.push(currentSlots);
    }
    return slots.sort((a, b) => a - b);
  })();

  $: visibleTiers = showAllSizes
    ? availableTiers
    : availableTiers.filter((t) => popularSlotsWithExtras.includes(t.slots));

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
    open = isOpen;
  }}
>
  <Dialog.Content class="max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Manage Slots</Dialog.Title>
      <Dialog.Description>
        {#if viewOnly}
          Based on your current plan, we recommend the following slot
          configuration. <a
            href="/{organization}/-/settings/billing"
            class="text-primary-500 hover:underline">Upgrade to Growth</a
          > to customize your slot allocation.
        {:else}
          All deployments are billed at ${SLOT_RATE_PER_HR}/slot/hr. Monthly
          estimates assume ~{HOURS_PER_MONTH} hours/month.
          {#if isRillManaged}
            Minimum {DEFAULT_MANAGED_SLOTS} slots for Rill-managed deployments.
          {:else}
            Minimum {DEFAULT_SELF_MANAGED_SLOTS} slots for self-managed OLAP deployments.
          {/if}
        {/if}
      </Dialog.Description>
    </Dialog.Header>

    <!-- Tier table -->
    <div class="tier-table">
      <div class="tier-header">
        <span class="tier-cell tier-cell-wide">Cluster Size</span>
        <span class="tier-cell">Slots</span>
        <span class="tier-cell">Est. $/mo</span>
      </div>
      <div class="tier-list">
        {#each visibleTiers as tier}
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
            <span class="tier-cell tier-cell-wide">
              {tier.instance}
            </span>
            <span class="tier-cell">
              {tier.slots}
              {#if tier.slots === currentSlots}
                <span class="current-badge">current</span>
              {/if}
              {#if tier.slots === minSlots}
                <span class="min-badge">min</span>
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

    <!-- Hibernate CTA -->
    <p class="hibernate-note">
      Want to stop billing entirely?
      <a
        href="/{organization}/{project}/-/settings"
        class="hibernate-link"
        on:click={() => (open = false)}
      >
        Hibernate this project
      </a>
      from the project settings page.
    </p>

    <div class="footer">
      <button class="cancel-btn" on:click={() => (open = false)}>
        Cancel
      </button>
      <button
        class="apply-btn"
        disabled={(viewOnly ? false : !hasChanged) || $updateProject.isPending}
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
  .min-badge {
    @apply text-[10px] text-fg-tertiary bg-surface-subtle px-1.5 py-0.5 rounded-full leading-none font-medium;
  }
  .show-all-btn {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0 mt-2;
  }
  .show-all-btn:hover {
    @apply text-primary-600;
  }
  .hibernate-note {
    @apply text-xs text-fg-tertiary mt-3 italic;
  }
  .hibernate-link {
    @apply text-primary-500 no-underline;
  }
  .hibernate-link:hover {
    @apply text-primary-600 underline;
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
