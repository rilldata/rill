<script lang="ts">
  import OptionsButton from "../column-profile/ContextButton.svelte";

  import MoreIcon from "../icons/MoreHorizontal.svelte";
  import { Menu, MenuItem } from "../menu";

  import { guidGenerator } from "../../util/guid";
  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";

  import Check from "@rilldata/web-local/lib/components/icons/Check.svelte";
  import Cancel from "@rilldata/web-local/lib/components/icons/Cancel.svelte";
  export let toggleFilterExcludeMode: () => void;
  export let filterExcludeMode: boolean;
  export let optionsMenuActive = false;

  const optionsButtonId = guidGenerator();
</script>

<WithTogglableFloatingElement
  bind:active={optionsMenuActive}
  let:toggleFloatingElement
>
  <OptionsButton
    id={optionsButtonId}
    tooltipText="leaderboard options"
    suppressTooltip={optionsMenuActive}
    on:click={toggleFloatingElement}
  >
    <MoreIcon />
  </OptionsButton>
  <Menu
    dark
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:item-select={toggleFloatingElement}
    slot="floating-element"
  >
    {#if filterExcludeMode}
      <MenuItem icon on:select={toggleFilterExcludeMode}>
        <Cancel slot="icon" size="20px" />
        Output excludes selections, click to include
      </MenuItem>
    {:else}
      <MenuItem icon on:select={toggleFilterExcludeMode}>
        <Check slot="icon" size="20px" />
        Output includes selections, click to exclude
      </MenuItem>
    {/if}
  </Menu>
</WithTogglableFloatingElement>
