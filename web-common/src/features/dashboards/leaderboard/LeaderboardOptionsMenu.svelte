<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import MoreIcon from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createEventDispatcher } from "svelte";

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
    marginClasses="ml-3"
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
