<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let dimensionName: string | undefined;
  export let isBeingCompared: boolean;
  export let toggleComparisonDimension: (dimensionName: string) => void;
</script>

<IconButton
  ariaLabel={m.leaderboard_toggle_breakdown({ name: dimensionName ?? "" })}
  onclick={(e) => {
    e.stopPropagation();
    if (dimensionName) toggleComparisonDimension(dimensionName);
  }}
>
  <Tooltip location="left" distance={8}>
    <Compare isColored={isBeingCompared} />
    <TooltipContent slot="tooltip-content">
      {isBeingCompared ? m.leaderboard_remove_comparison() : m.leaderboard_compare()}
    </TooltipContent>
  </Tooltip>
</IconButton>
