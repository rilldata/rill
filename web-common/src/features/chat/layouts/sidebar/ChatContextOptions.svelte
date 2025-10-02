<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Button } from "@rilldata/web-common/components/button";
  import SettingsSlider from "@rilldata/web-common/components/icons/SettingsSlider.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import type { StateManagers } from "../../../dashboards/state-managers/state-managers";
  import { writable } from "svelte/store";

  export let stateManagers: StateManagers | undefined = undefined;
  export let includeFilters: boolean = true;
  export let includeTimeRange: boolean = true;
  export let onOptionsChange: (options: {
    includeFilters: boolean;
    includeTimeRange: boolean;
  }) => void = () => {};

  // Only show context options if we have dashboard state
  $: hasDashboardContext = !!stateManagers;

  function toggleFilters() {
    includeFilters = !includeFilters;
    onOptionsChange({ includeFilters, includeTimeRange });
  }

  function toggleTimeRange() {
    includeTimeRange = !includeTimeRange;
    onOptionsChange({ includeFilters, includeTimeRange });
  }
</script>

{#if hasDashboardContext}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button type="ghost" builders={[builder]} class="h-7 w-7 p-0">
        <SettingsSlider size="14px" />
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="w-56">
      <DropdownMenu.Label class="text-xs font-medium text-gray-700">
        Dashboard Context
      </DropdownMenu.Label>
      <DropdownMenu.Separator />

      <DropdownMenu.CheckboxItem
        bind:checked={includeFilters}
        on:click={toggleFilters}
        class="flex items-center gap-2"
      >
        <div class="flex items-center justify-center w-4 h-4">
          {#if includeFilters}
            <Check size="12px" />
          {/if}
        </div>
        <div class="flex-1">
          <div class="text-sm font-medium">Include Filters</div>
          <div class="text-xs text-gray-500">
            Use current dashboard filters in queries
          </div>
        </div>
      </DropdownMenu.CheckboxItem>

      <DropdownMenu.CheckboxItem
        bind:checked={includeTimeRange}
        on:click={toggleTimeRange}
        class="flex items-center gap-2"
      >
        <div class="flex items-center justify-center w-4 h-4">
          {#if includeTimeRange}
            <Check size="12px" />
          {/if}
        </div>
        <div class="flex-1">
          <div class="text-sm font-medium">Include Time Range</div>
          <div class="text-xs text-gray-500">
            Use current time period in queries
          </div>
        </div>
      </DropdownMenu.CheckboxItem>

      <DropdownMenu.Separator />
      <DropdownMenu.Item disabled class="text-xs text-gray-500">
        Context helps the AI understand your current dashboard state
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
