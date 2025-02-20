<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/shadcn/DashboardVisibilityDropdown.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let exploreName: string;

  const {
    selectors: {
      measures: { leaderboardMeasureName, getMeasureByName, visibleMeasures },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextCol: { setContextColumn },
      setLeaderboardMeasureName,
    },
  } = getStateManagers();

  let active = false;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);
  $: console.log("activeLeaderboardMeasure: ", activeLeaderboardMeasure);
  $: console.log("measures: ", measures);

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  $: visibleDimensionsNames = $visibleDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allDimensionNames = $allDimensions
    .map(({ name }) => name)
    .filter(isDefined);

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    metricsExplorer?.leaderboardContextColumn ===
      LeaderboardContextColumn.PERCENT
  ) {
    setContextColumn(LeaderboardContextColumn.HIDDEN);
  }

  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  let disabled = false;
</script>

<div>
  {#if measures.length && activeLeaderboardMeasure}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-1"
      style:max-width="450px"
    >
      <DashboardVisibilityDropdown
        category="Dimensions"
        tooltipText="Choose dimensions to display"
        onSelect={(name) => toggleDimensionVisibility(allDimensionNames, name)}
        selectableItems={$allDimensions.map(({ name, displayName }) => ({
          name: name || "",
          label: displayName || name || "",
        }))}
        selectedItems={visibleDimensionsNames}
        onToggleSelectAll={() => {
          toggleDimensionVisibility(allDimensionNames);
        }}
      />

      <!-- TODO: move to a separate component -->
      <DropdownMenu.Root
        closeOnItemClick={false}
        typeahead={false}
        bind:open={active}
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
              label={activeLeaderboardMeasure.displayName ||
                activeLeaderboardMeasure.name}
              on:click
            >
              <div
                class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
              >
                Showing <strong
                  >{`${activeLeaderboardMeasure.displayName || activeLeaderboardMeasure.name}`}</strong
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

            <DropdownMenu.Content>
              {#each measures as measure (measure.name)}
                <DropdownMenu.Item class="text-[12px]">
                  <div class="flex flex-col">
                    <div
                      class:font-bold={$leaderboardMeasureName === measure.name}
                    >
                      {measure.displayName || measure.name}
                    </div>

                    <p class="ui-copy-muted" style:font-size="11px">
                      {measure.description}
                    </p>
                  </div>
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>

            <div
              slot="tooltip-content"
              transition:fly={{ duration: 300, y: 4 }}
            >
              <TooltipContent maxWidth="400px">
                Choose a measure to filter by
              </TooltipContent>
            </div>
          </Tooltip>
        </DropdownMenu.Trigger>
      </DropdownMenu.Root>

      <!-- <Select.Root
        bind:open={active}
        selected={{ value: activeLeaderboardMeasure.name, label: "" }}
        items={measures.map((measure) => ({
          value: measure.name ?? "",
          label: measure.displayName || measure.name,
        }))}
        onSelectedChange={(newSelection) => {
          if (!newSelection?.value) return;
          setLeaderboardMeasureName(newSelection.value);
        }}
      >
        <Select.Trigger class="outline-none border-none w-fit  px-0 gap-x-0.5">
          <Button type="text" label="Select a measure to filter by">
            <span class="truncate text-gray-700 hover:text-inherit">
              Showing <b>
                {activeLeaderboardMeasure?.displayName ||
                  activeLeaderboardMeasure.name}
              </b>
            </span>
          </Button>
        </Select.Trigger>

        <Select.Content
          sameWidth={false}
          align="start"
          class="max-h-80 overflow-y-auto min-w-44"
        >
          {#each measures as measure (measure.name)}
            <Select.Item
              value={measure.name}
              label={measure.displayName || measure.name}
              class="text-[12px]"
            >
              <div class="flex flex-col">
                <div class:font-bold={$leaderboardMeasureName === measure.name}>
                  {measure.displayName || measure.name}
                </div>

                <p class="ui-copy-muted" style:font-size="11px">
                  {measure.description}
                </p>
              </div>
            </Select.Item>
          {/each}
        </Select.Content>
      </Select.Root> -->
    </div>
  {/if}
</div>
