<script lang="ts">
  import MoreIcon from "../icons/MoreHorizontal.svelte";
  import { Menu, MenuItem } from "../menu";
  import IconButton from "../button/IconButton.svelte";

  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";
  import { createEventDispatcher } from "svelte";

  import Check from "@rilldata/web-local/lib/components/icons/Check.svelte";
  import Cancel from "@rilldata/web-local/lib/components/icons/Cancel.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  export let filterExcludeMode: boolean;
  export let optionsMenuActive = false;

  const dispatch = createEventDispatcher();
  let toggleFilterExcludeMode = () => dispatch("toggle-filter-exclude-mode");
</script>

<WithTogglableFloatingElement
  bind:active={optionsMenuActive}
  let:toggleFloatingElement
>
  <IconButton on:click={toggleFloatingElement}>
    <Tooltip location="right" distance={8} suppress={optionsMenuActive}>
      <MoreIcon />
      <TooltipContent slot="tooltip-content">leaderboard options</TooltipContent
      >
    </Tooltip>
  </IconButton>
  <Menu
    dark
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:item-select={toggleFloatingElement}
    slot="floating-element"
  >
    {#if filterExcludeMode}
      <MenuItem icon on:select={toggleFilterExcludeMode}>
        <Check slot="icon" size="20px" />
        click to include selected values
      </MenuItem>
    {:else}
      <MenuItem icon on:select={toggleFilterExcludeMode}>
        <Cancel slot="icon" size="20px" />
        click to exclude selected values
      </MenuItem>
    {/if}
  </Menu>
</WithTogglableFloatingElement>
