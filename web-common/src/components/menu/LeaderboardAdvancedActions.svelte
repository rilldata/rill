<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

  const ctx = getStateManagers();
  const {
    exploreName,
    selectors: {
      leaderboard: { leaderboardShowContextForAllMeasures },
    },
  } = ctx;

  const timeControlsStore = useTimeControlStore(ctx);
  $: ({ showTimeComparison } = $timeControlsStore);

  export let isOpen: boolean;
  export let toggle: () => void;

  // Ensure time comparison is enabled before toggling the leaderboard show context for all measures
  function ensureTimeComparisonEnabled() {
    if (!showTimeComparison) {
      metricsExplorerStore.displayTimeComparison($exploreName, true);
    }
  }
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger>
    <IconButton rounded active={isOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </PopoverTrigger>
  <PopoverContent
    align="start"
    side="bottom"
    class="flex flex-row items-center justify-between gap-x-2 w-[286px] px-3.5 py-2.5"
  >
    <span>Show context for all measures</span>
    <Switch
      theme
      checked={$leaderboardShowContextForAllMeasures}
      onCheckedChange={() => {
        ensureTimeComparisonEnabled();

        // Avoid race condition
        setTimeout(() => toggle(), 0);
      }}
      small
    />
  </PopoverContent>
</Popover>
