<script lang="ts">
  import MoreIcon from "../icons/MoreHorizontal.svelte";
  import { Menu } from "../menu";
  import { Switch } from "@rilldata/web-local/lib/components/button";

  import IconButton from "../button/IconButton.svelte";

  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";
  import { createEventDispatcher } from "svelte";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let filterExcludeMode: boolean;
  export let optionsMenuActive = false;

  const dispatch = createEventDispatcher();
  let toggleFilterMode = () => dispatch("toggle-filter-mode");
</script>

<WithTogglableFloatingElement
  bind:active={optionsMenuActive}
  let:toggleFloatingElement
>
  <IconButton
    on:click={(e) => {
      e.stopPropagation();
      toggleFloatingElement();
    }}
  >
    <Tooltip location="right" distance={8} suppress={optionsMenuActive}>
      <MoreIcon />
      <TooltipContent slot="tooltip-content">Leaderboard options</TooltipContent
      >
    </Tooltip>
  </IconButton>
  <Menu
    dark
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:item-select={toggleFloatingElement}
    minWidth="200px"
    slot="floating-element"
  >
    <Switch
      showBgOnHover={false}
      on:click={() => toggleFilterMode()}
      checked={filterExcludeMode}
    >
      <span class="text-white" slot="left">Exclude selected items</span>
    </Switch>
  </Menu>
</WithTogglableFloatingElement>
