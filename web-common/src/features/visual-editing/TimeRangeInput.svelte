<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    PERIOD_TO_DATE_RANGES,
    LATEST_WINDOW_TIME_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
    DEFAULT_TIME_RANGES,
  } from "@rilldata/web-common/lib/time/config";

  const ranges = [
    ...Object.keys(LATEST_WINDOW_TIME_RANGES),
    ...Object.keys(PERIOD_TO_DATE_RANGES),
    ...Object.keys(PREVIOUS_COMPLETE_DATE_RANGES),
  ];

  const defaultSet = new Set(ranges);

  export let selectedItems: Set<string>;
  export let onSelectDefault: (defaults: string[]) => void;
  export let onSelectCustomItem: (item: string) => void;
  export let restoreDefaults: (defaults: string[]) => void;

  let open = false;
  let searchValue = "";

  $: hasDefaultsSelected =
    defaultSet.size === selectedItems.size &&
    defaultSet.isSubsetOf(selectedItems);

  $: mode = hasDefaultsSelected ? "default" : "custom";

  // $: filteredNonDefaults = allNonDefaults.filter((item) =>
  //   item.toLowerCase().includes(searchValue.toLowerCase()),
  // );

  $: selected = mode === "custom" ? 1 : 0;

  $: filteredItems = ranges.filter(
    (item) =>
      !selectedItems.has(item) &&
      item.toLowerCase().includes(searchValue.toLowerCase()),
  );

  // function onToggleSelectAll() {}
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    capitalize={false}
    label="Available time ranges"
    id="visual-explore-range"
  />
  <FieldSwitcher
    fields={["Default", "Custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "Custom") {
        mode = "custom";
      } else if (field === "Default") {
        onSelectDefault(ranges);
      }
    }}
  />

  {#if mode === "custom"}
    <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
      <DropdownMenu.Trigger asChild let:builder>
        <button
          use:builder.action
          {...builder}
          class:open
          class="flex px-3 gap-x-2 h-8 max-w-full items-center text-sm border-gray-300 border rounded-[2px] break-all overflow-hidden"
        >
          {selectedItems.size} of {ranges.length}
          <CaretDownIcon
            size="12px"
            className="!fill-gray-600 ml-auto flex-none"
          />
        </button>
      </DropdownMenu.Trigger>

      <DropdownMenu.Content sameWidth class="p-0">
        <div class="p-3 pb-1">
          <Search bind:value={searchValue} autofocus={false} />
        </div>
        <div class="max-h-64 overflow-y-auto">
          {#each selectedItems as item (item)}
            <DropdownMenu.CheckboxItem
              checked
              on:click={() => onSelectCustomItem(item)}
            >
              {DEFAULT_TIME_RANGES[item].label}
            </DropdownMenu.CheckboxItem>
          {/each}

          {#if selectedItems.size > 0 && filteredItems.length > 0}
            <DropdownMenu.Separator />
          {/if}

          {#each filteredItems as item (item)}
            <DropdownMenu.CheckboxItem
              on:click={() => onSelectCustomItem(item)}
            >
              {DEFAULT_TIME_RANGES[item].label}
            </DropdownMenu.CheckboxItem>
          {/each}
        </div>

        <footer>
          {#if !hasDefaultsSelected}
            <Button on:click={() => restoreDefaults(ranges)} type="text">
              Restore defaults
            </Button>
          {/if}
          <!-- <Button on:click={onToggleSelectAll} type="plain">
              {#if selectedItems.size === DEFAULT_TIMEZONES.length}
                Deselect all
              {:else}
                Select all
              {/if}
            </Button> -->
        </footer>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>

<style lang="postcss">
  .open {
    @apply ring-2 ring-primary-100 border-primary-600;
  }

  footer {
    @apply mt-1;
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }
</style>
