<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

  export let tooltipText: string;
  export let searchText = "";
  export let measures: MetricsViewSpecMeasureV2[];
  export let firstMeasure: MetricsViewSpecMeasureV2;
  export let selectedMeasureNames: string[] = [];
  export let onToggle: (name: string) => void;
  export let onSelectAll: () => void;

  let active = false;
  let disabled = false;

  function filterMeasures(searchText: string) {
    return measures.filter((item) =>
      ((item.displayName || item.name) ?? "")
        .toLowerCase()
        .includes(searchText.toLowerCase()),
    );
  }

  function toggleMeasure(name: string) {
    if (!name) return;
    onToggle(name);
  }

  function getMeasureDisplayText(measureName: string) {
    const measure = measures.find((m) => m.name === measureName);
    return measure?.displayName || measure?.name || measureName;
  }

  $: filteredMeasures = filterMeasures(searchText);

  $: showingMeasuresText =
    selectedMeasureNames.length > 1
      ? `${selectedMeasureNames.length} measures`
      : getMeasureDisplayText(selectedMeasureNames[0]);

  $: allSelected = selectedMeasureNames.length === measures.length;
</script>

<DropdownMenu.Root
  closeOnItemClick={false}
  typeahead={false}
  bind:open={active}
  onOpenChange={(open) => {
    if (!open) searchText = "";
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Button
        builders={[builder]}
        type="text"
        label={firstMeasure.displayName || firstMeasure.name}
        on:click
      >
        <div
          class="flex items-center gap-x-1 px-1 text-gray-700 hover:text-inherit font-normal"
        >
          Showing <strong>{showingMeasuresText}</strong>
          <span
            class="transition-transform"
            class:hidden={disabled}
            class:-rotate-180={active}
          >
            <CaretDownIcon />
          </span>
        </div>
      </Button>

      <DropdownMenu.Content
        align="start"
        class="flex flex-col max-h-96 w-72 overflow-hidden p-0"
      >
        <div class="px-3 pt-3 pb-1">
          <Search
            bind:value={searchText}
            label="Search measures"
            showBorderOnFocus={false}
          />
        </div>

        <div class="px-1 pb-1 overflow-y-auto">
          {#if filteredMeasures.length}
            {#each filteredMeasures as measure (measure.name)}
              <DropdownMenu.CheckboxItem
                class="text-[12px]"
                checked={measure.name
                  ? selectedMeasureNames.includes(measure.name)
                  : false}
                onCheckedChange={() => {
                  if (measure.name) toggleMeasure(measure.name);
                }}
              >
                <div class="flex flex-col max-w-full">
                  <div>
                    {measure.displayName || measure.name}
                  </div>

                  <p
                    class="ui-copy-muted truncate"
                    title={measure.description}
                    style:font-size="11px"
                  >
                    {measure.description}
                  </p>
                </div>
              </DropdownMenu.CheckboxItem>
            {/each}
          {:else}
            <div class="ui-copy-disabled text-center p-2 w-full">
              no results
            </div>
          {/if}
        </div>

        <footer>
          <Button on:click={onSelectAll} type="plain">
            {#if allSelected}
              Deselect all
            {:else}
              Select all
            {/if}
          </Button>
        </footer>
      </DropdownMenu.Content>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>
</DropdownMenu.Root>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
