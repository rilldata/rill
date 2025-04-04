<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

  export let leaderboardSortByMeasureName: string;
  export let activeLeaderboardMeasure: MetricsViewSpecMeasureV2 | undefined;
  export let measures: MetricsViewSpecMeasureV2[];
  export let setLeaderboardSortByMeasureName: (name: string) => void;

  let active = false;
</script>

{#if activeLeaderboardMeasure}
  <Select.Root
    bind:open={active}
    selected={{ value: activeLeaderboardMeasure.name, label: "" }}
    items={measures.map((measure) => ({
      value: measure.name ?? "",
      label: measure.displayName || measure.name,
    }))}
    onSelectedChange={(newSelection) => {
      if (!newSelection?.value) return;
      setLeaderboardSortByMeasureName(newSelection.value);
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
            <div
              class:font-bold={leaderboardSortByMeasureName === measure.name}
            >
              {measure.displayName || measure.name}
            </div>

            <p class="ui-copy-muted" style:font-size="11px">
              {measure.description}
            </p>
          </div>
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
{/if}
