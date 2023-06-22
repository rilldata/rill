<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import MoreIcon from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { createEventDispatcher } from "svelte";
  import { tick } from "svelte/internal";
  import ContextButton from "../column-profile/ContextButton.svelte";
  import TableHeader from "./TableHeader.svelte";

  export let index: number;
  export let menuLabel: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  let rowHovered = false;

  const rowMouseEnter = () => {
    rowHovered = true;
  };
  const rowMouseLeave = async () => {
    setTimeout(() => {
      rowHovered = false;
    }, 5);
  };
  let menuContainerHovered = false;
  const menuContainerEnter = () => {
    menuContainerHovered = true;
  };

  const menuContainerLeave = async () => {
    await tick();
    menuContainerHovered = false;
  };

  const contextButtonId = guidGenerator();

  let contextMenuActive = false;

  $: rowActive = rowHovered || menuContainerHovered;
</script>

<tr
  class="
        hover:bg-gray-100
        {rowActive && 'bg-gray-100'}
    "
  on:focus={() => (rowHovered = true)}
  on:mouseenter={rowMouseEnter}
  on:mouseleave={rowMouseLeave}
>
  <TableHeader position="left">
    <span style={rowActive ? "visibility:hidden" : ""}>{index + 1}</span>
    {#if rowActive || contextMenuActive}
      <div
        style="position:absolute; top:50%; left:50%; width: 0px; height: 0px;"
        on:mouseenter={menuContainerEnter}
        on:mouseleave={menuContainerLeave}
        class="bg-gray-200"
      >
        <div style="position:absolute; top:-8px; left:-8px">
          <WithTogglableFloatingElement
            bind:active={contextMenuActive}
            let:toggleFloatingElement
          >
            <ContextButton
              id={contextButtonId}
              tooltipText=""
              suppressTooltip={true}
              on:click={toggleFloatingElement}
              label={menuLabel}
            >
              <MoreIcon />
            </ContextButton>
            <Menu
              dark
              on:escape={toggleFloatingElement}
              on:click-outside={toggleFloatingElement}
              on:item-select={toggleFloatingElement}
              slot="floating-element"
            >
              <MenuItem icon on:select={() => dispatch("delete")}>
                <Cancel slot="icon" />
                Delete row</MenuItem
              >
            </Menu>
          </WithTogglableFloatingElement>
        </div>
      </div>
    {/if}
  </TableHeader>
  <slot />
</tr>
