<script lang="ts">
  import { tick } from "svelte/internal";
  import { createEventDispatcher } from "svelte";

  import Portal from "$lib/components/Portal.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";

  import { guidGenerator } from "$lib/util/guid";
  import { onClickOutside } from "$lib/util/on-click-outside";

  const dispatch = createEventDispatcher();

  let rowHovered = false;

  let rowRect: DOMRect | undefined;

  let menuX: number;
  let menuY: number;

  const rowMouseEnter = (e) => {
    rowRect = (<HTMLTableRowElement>e.target).getBoundingClientRect();
    rowHovered = true;
  };
  const rowMouseLeave = async () => {
    setTimeout(() => {
      rowHovered = false;
    }, 5);
  };

  const MENU_CONTAINER_WIDTH = 23;
  $: menuContainerStyle =
    rowRect === undefined
      ? ""
      : `z-index:50;
position: fixed;
top: ${rowRect.top}px;
left: ${rowRect.left - MENU_CONTAINER_WIDTH}px;
width: ${MENU_CONTAINER_WIDTH}px;
height: ${rowRect.height}px;
display: flex;
justify-content: center;
align-items: center;
`;

  let menuContainerHovered = false;
  const menuContainerEnter = (e) => {
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
  <slot />
</tr>

{#if rowActive}
  <Portal>
    <div
      style={menuContainerStyle}
      on:mouseenter={menuContainerEnter}
      on:mouseleave={menuContainerLeave}
      class="self-center"
    >
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
  </Portal>
{/if}

{#if contextMenuOpen}
  <div bind:this={contextMenu}>
    <FloatingElement
      relationship="mouse"
      target={{ x: menuX, y: menuY }}
      location="left"
      alignment="start"
    >
      <Menu on:escape={closeContextMenu} on:item-select={closeContextMenu}>
        <MenuItem on:select={() => dispatch("delete")}>delete row</MenuItem>
      </Menu>
    </FloatingElement>
  </div>
{/if}
