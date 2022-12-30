<script>
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let suppressTooltips = false;
  export let active;
</script>

<div class="flex items-start w-full">
  <WithTogglableFloatingElement
    location="left"
    distance={16}
    let:toggleFloatingElement
    bind:active
  >
    <Tooltip
      location="bottom"
      distance={8}
      suppress={active || suppressTooltips}
    >
      <button
        style:font-size="20px"
        style:height="36px"
        class="drag-handle opacity-0 focus:opacity-100"
        on:click={(event) => {
          toggleFloatingElement();
        }}
        on:mousedown
      >
        <DragHandle />
      </button>
      <TooltipContent slot="tooltip-content"
        >move entry, delete ...</TooltipContent
      >
    </Tooltip>
    <Menu
      on:escape={toggleFloatingElement}
      on:click-outside={toggleFloatingElement}
      slot="floating-element"
      on:item-select={toggleFloatingElement}
    >
      <slot name="menu-items" />
    </Menu>
  </WithTogglableFloatingElement>
  <slot />
</div>

<style>
  :global(.container:hover .drag-handle) {
    opacity: 1;
  }
</style>
