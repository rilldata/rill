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
  export let disabled = false;
  export let searchText = "";
  export let measures: MetricsViewSpecMeasureV2[];
  export let activeMeasure: MetricsViewSpecMeasureV2;
  export let leaderboardMeasureName: string;
  export let onSelect: (name: string) => void;

  let active = false;

  function filterMeasures(searchText: string) {
    return measures.filter((item) =>
      ((item.displayName || item.name) ?? "")
        .toLowerCase()
        .includes(searchText.toLowerCase()),
    );
  }

  $: filteredMeasures = filterMeasures(searchText);
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
        label={activeMeasure.displayName || activeMeasure.name}
        on:click
      >
        <div
          class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
        >
          Showing <strong
            >{`${activeMeasure.displayName || activeMeasure.name}`}</strong
          >
          <span
            class="transition-transform"
            class:hidden={disabled}
            class:-rotate-180={active}
          >
            <CaretDownIcon />
          </span>
        </div>
      </Button>

      <!-- TODO: select or deselect all when internal state supports multi-select -->
      <DropdownMenu.Content
        align="start"
        class="flex flex-col max-h-96 w-72 overflow-hidden"
      >
        <div class="pb-1">
          <Search
            bind:value={searchText}
            label="Search measures"
            showBorderOnFocus={false}
          />
        </div>

        <!-- TODO: checkbox -->
        {#if filteredMeasures.length}
          {#each filteredMeasures as measure (measure.name)}
            <DropdownMenu.Item
              class="text-[12px]"
              on:click={() => {
                if (measure.name) onSelect(measure.name);
              }}
            >
              <div class="flex flex-col">
                <div class:font-bold={leaderboardMeasureName === measure.name}>
                  {measure.displayName || measure.name}
                </div>

                <p class="ui-copy-muted" style:font-size="11px">
                  {measure.description}
                </p>
              </div>
            </DropdownMenu.Item>
          {/each}
        {:else}
          <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
        {/if}
      </DropdownMenu.Content>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>
</DropdownMenu.Root>
