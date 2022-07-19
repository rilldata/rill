<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { tick } from "svelte/internal";

  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import TableHeader from "./TableHeader.svelte";

  import { guidGenerator } from "$lib/util/guid";
  import { onClickOutside } from "$lib/util/on-click-outside";

  export let index: number;
  const dispatch = createEventDispatcher();

  let rowHovered = false;

  let menuX: number;
  let menuY: number;

  const rowMouseEnter = (e) => {
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

  let contextMenuOpen = false;
  const closeContextMenu = () => {
    contextMenuOpen = false;
  };
  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }
  let contextMenu: any;

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
    {#if rowActive}
      <div
        style="position:absolute; top:50%; left:50%; width: 0px; height: 0px;"
        on:mouseenter={menuContainerEnter}
        on:mouseleave={menuContainerLeave}
        class="bg-gray-200"
      >
        <div style="position:absolute; top:-8px; left:-8px">
          <ContextButton
            id={contextButtonId}
            tooltipText=""
            suppressTooltip={true}
            on:click={async (event) => {
              contextMenuOpen = !contextMenuOpen;
              menuX = event.clientX;
              menuY = event.clientY;
              if (!clickOutsideListener) {
                await tick();
                clickOutsideListener = onClickOutside(() => {
                  contextMenuOpen = false;
                }, contextMenu);
              }
            }}
          >
            <MoreIcon />
          </ContextButton>
        </div>
      </div>
    {/if}
  </TableHeader>
  <slot />
</tr>

{#if contextMenuOpen}
  <div bind:this={contextMenu}>
    <FloatingElement
      relationship="mouse"
      target={{ x: menuX, y: menuY }}
      location="left"
      alignment="start"
    >
      <Menu dark on:escape={closeContextMenu} on:item-select={closeContextMenu}>
        <MenuItem on:select={() => dispatch("delete")}>delete row</MenuItem>
      </Menu>
    </FloatingElement>
  </div>
{/if}
