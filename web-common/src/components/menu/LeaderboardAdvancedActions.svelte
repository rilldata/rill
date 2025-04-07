<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";

  export let isOpen: boolean;
  export let leaderboardShowAllMeasures: boolean;
  export let setLeaderboardShowAllMeasures: (value: boolean) => void;
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger>
    <IconButton rounded active={isOpen} compact>
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
      checked={leaderboardShowAllMeasures}
      on:click={() => {
        setLeaderboardShowAllMeasures(!leaderboardShowAllMeasures);
      }}
      small
    />
  </PopoverContent>
</Popover>
