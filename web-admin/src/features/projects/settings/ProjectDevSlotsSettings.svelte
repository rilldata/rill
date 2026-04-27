<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseUpdateProjectError } from "@rilldata/web-admin/features/projects/settings/errors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import {
    ALL_SLOTS,
    SLOT_TIERS,
  } from "@rilldata/web-admin/features/projects/status/overview/slots-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";

  let { organization, project }: { organization: string; project: string } =
    $props();

  const updateProjectMutation = createAdminServiceUpdateProject();
  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );

  let currentDevSlots = $derived(
    Number($projectResp.data?.project?.devSlots) || 0,
  );

  let dropdownOpen = $state(false);

  function formatSize(slots: number): { units: string; instance: string } {
    const tier = SLOT_TIERS.find((t) => t.slots === slots);
    const instance = tier?.instance ?? `${slots * 4}GiB / ${slots}vCPU`;
    const unitLabel = slots === 1 ? "Compute unit" : "Compute units";
    return { units: `${slots} ${unitLabel}`, instance };
  }

  // Dev deployments may use a single slot, which isn't in ALL_SLOTS (that list
  // starts at 2 for production sizing).
  let slotOptions = $derived(
    (() => {
      const slots = new Set<number>([1, ...ALL_SLOTS]);
      if (currentDevSlots > 0) slots.add(currentDevSlots);
      return [...slots].sort((a, b) => a - b);
    })(),
  );

  let current = $derived(
    currentDevSlots > 0 ? formatSize(currentDevSlots) : null,
  );

  async function handleSelect(slots: number) {
    if (slots === currentDevSlots) return;
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project,
        data: { devSlots: String(slots) },
      });
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });
      const { units, instance } = formatSize(slots);
      eventBus.emit("notification", {
        message: `Default dev cluster size updated to ${units} — ${instance}`,
      });
    } catch (err) {
      const parsed = parseUpdateProjectError(err as AxiosError<RpcStatus>);
      eventBus.emit("notification", {
        message: parsed.message || "Failed to update dev cluster size",
        type: "error",
      });
    }
  }
</script>

<SettingsContainer title="Development deployments">
  <p>
    Sets the default cluster size allocated to new development deployments.
    Development deployments are ephemeral — they spin up when a user opens a
    branch editor and expire after 6 hours of inactivity.
  </p>

  <div class="field">
    <div class="label-row">
      <span class="label">Default cluster size</span>
      <Tooltip location="right" alignment="middle" distance={8}>
        <div class="text-fg-tertiary flex items-center">
          <InfoCircle size="12px" />
        </div>
        <TooltipContent slot="tooltip-content">
          1 Compute unit = 4 GiB RAM, 1 vCPU
        </TooltipContent>
      </Tooltip>
    </div>
    <div class="row">
      <span class="value">
        {#if $updateProjectMutation.isPending}
          <span class="text-fg-tertiary">Updating…</span>
        {:else if current}
          <span class="units">{current.units}</span>
          <span class="separator">—</span>
          <span class="instance">{current.instance}</span>
        {:else}
          <span class="text-fg-tertiary">Not set</span>
        {/if}
      </span>

      <DropdownMenu.Root bind:open={dropdownOpen}>
        <DropdownMenu.Trigger
          class="change-trigger {dropdownOpen ? 'open' : ''}"
          disabled={$updateProjectMutation.isPending}
        >
          <span>Select cluster size</span>
          {#if dropdownOpen}
            <CaretUpIcon size="12px" />
          {:else}
            <CaretDownIcon size="12px" />
          {/if}
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end" class="w-[260px]">
          {#each slotOptions as slots (slots)}
            {@const opt = formatSize(slots)}
            <DropdownMenu.Item
              class="flex flex-col items-start gap-0.5 {slots ===
              currentDevSlots
                ? 'bg-surface-subtle'
                : ''}"
              onclick={() => handleSelect(slots)}
            >
              <span class="text-xs font-medium text-fg-primary">
                {opt.units}
              </span>
              <span class="text-[11px] text-fg-tertiary">
                {opt.instance}
              </span>
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  </div>
</SettingsContainer>

<style lang="postcss">
  p {
    @apply text-sm text-fg-tertiary;
  }

  .field {
    @apply mt-4 flex flex-col gap-y-1;
  }

  .label-row {
    @apply flex items-center gap-x-1;
  }

  .label {
    @apply text-sm font-medium text-fg-primary;
  }

  .row {
    @apply flex items-center justify-between gap-4 py-1;
  }

  .value {
    @apply text-sm text-fg-primary flex items-baseline gap-x-2;
  }

  .units {
    @apply font-medium;
  }

  .separator {
    @apply text-fg-tertiary;
  }

  .instance {
    @apply text-fg-secondary;
  }

  :global(.change-trigger) {
    @apply flex items-center gap-1 rounded-sm border px-2.5 py-1 text-sm text-fg-primary transition-colors;
    @apply hover:bg-surface-hover;
  }

  :global(.change-trigger.open) {
    @apply bg-surface-hover;
  }

  :global(.change-trigger[disabled]) {
    @apply opacity-50 cursor-not-allowed;
  }
</style>
