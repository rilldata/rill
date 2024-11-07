<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
  import ZoneDisplay from "../dashboards/time-controls/super-pill/components/ZoneDisplay.svelte";

  const defaultSet = new Set(DEFAULT_TIMEZONES);
  const allNonDefaults = Intl.supportedValuesOf("timeZone");

  export let selectedItems: Set<string>;
  export let keyNotSet: boolean;
  export let onSelectDefault: () => void;
  export let onSelectCustomItem: (item: string) => void;
  export let restoreDefaults: () => void;

  let open = false;
  let searchValue = "";

  $: hasDefaultsSelected =
    keyNotSet ||
    (defaultSet.size === selectedItems.size &&
      defaultSet.isSubsetOf(selectedItems));

  $: mode = hasDefaultsSelected ? "default" : "custom";

  $: filteredNonDefaults = allNonDefaults.filter((item) =>
    item.toLowerCase().includes(searchValue.toLowerCase()),
  );

  $: selected = mode === "custom" ? 1 : 0;

  $: filteredItems = DEFAULT_TIMEZONES.filter(
    (item) =>
      !selectedItems.has(item) &&
      item.toLowerCase().includes(searchValue.toLowerCase()),
  );

  // function onToggleSelectAll() {}
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    capitalize={false}
    label="Available time zones"
    id="visual-explore-zone"
  />
  <FieldSwitcher
    fields={["Default", "Custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "Custom") {
        mode = "custom";
      } else if (field === "Default") {
        onSelectDefault();
        mode = "default";
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
          {selectedItems.size} time zones
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
          {#if searchValue}
            {#each filteredNonDefaults as item (item)}
              <DropdownMenu.CheckboxItem
                checked={selectedItems.has(item)}
                on:click={() => onSelectCustomItem(item)}
              >
                <ZoneDisplay iana={item} />
              </DropdownMenu.CheckboxItem>
            {:else}
              {#if searchValue}
                <div class="ui-copy-disabled text-center p-2 w-full">
                  no results
                </div>
              {/if}
            {/each}
          {:else}
            {#each selectedItems as item (item)}
              <DropdownMenu.CheckboxItem
                checked
                on:click={() => onSelectCustomItem(item)}
              >
                <ZoneDisplay iana={item} />
              </DropdownMenu.CheckboxItem>
            {/each}

            {#if selectedItems.size > 0 && filteredItems.length > 0}
              <DropdownMenu.Separator />
            {/if}

            {#each filteredItems as item (item)}
              <DropdownMenu.CheckboxItem
                on:click={() => onSelectCustomItem(item)}
              >
                <ZoneDisplay iana={item} />
              </DropdownMenu.CheckboxItem>
            {/each}
          {/if}
        </div>

        <footer>
          {#if !hasDefaultsSelected}
            <Button on:click={restoreDefaults} type="text">
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
