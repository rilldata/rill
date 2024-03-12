<script lang="ts">
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import ExpanderButton from "@rilldata/web-common/components/column-profile/ExpanderButton.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { createCommandClickAction } from "../../lib/actions/command-click-action";
  import { createShiftClickAction } from "../../lib/actions/shift-click-action";
  import { emitNavigationTelemetry } from "./navigation-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let name: string;
  export let href: string;
  export let open = false;
  export let expandable = false;
  export let showContextMenu = true;

  const { commandClickAction } = createCommandClickAction();
  const { shiftClickAction } = createShiftClickAction();

  let showDetails = false;
  let contextMenuOpen = false;
  let mousedown = false;

  function onShowDetails() {
    showDetails = !showDetails;
  }

  async function shiftClickHandler() {
    await navigator.clipboard.writeText(name);
    notifications.send({ message: `copied "${name}" to clipboard` });
  }

  function handleMouseDown() {
    mousedown = true;
    function handleMouseUp() {
      mousedown = false;
      window.removeEventListener("mouseup", handleMouseUp);
    }
    window.addEventListener("mouseup", handleMouseUp);
  }
</script>

<li
  aria-label={name}
  class="entry group w-full flex gap-x-2"
  class:open={open || mousedown}
>
  {#if expandable}
    <ExpanderButton rotated={showDetails} on:click={onShowDetails} />
  {/if}

  <Tooltip location="right" alignment="start" distance={48} activeDelay={0}>
    <svelte:element
      this={open && expandable ? "button" : "a"}
      role="link"
      class="clickable-text"
      class:expandable
      class:open
      tabindex={open ? -1 : 0}
      href={open ? undefined : href}
      use:shiftClickAction
      use:commandClickAction
      on:command-click
      on:mousedown={handleMouseDown}
      on:shift-click={shiftClickHandler}
      on:click={() => {
        if (!open) {
          emitNavigationTelemetry(href);
          return;
        }
        if (expandable) onShowDetails();
      }}
    >
      {#if $$slots["icon"]}
        <div class="text-gray-400" style:width="1em" style:height="1em">
          <slot name="icon" />
        </div>
      {/if}
      <div class:truncate={!$$slots["name"]} class="w-full">
        {#if $$slots["name"]}
          <slot name="name" />
        {:else}
          {name}
        {/if}
      </div>
    </svelte:element>
    <!-- if tooltip content is present in a slot, render the tooltip -->
    <div slot="tooltip-content" class:hidden={!$$slots["tooltip-content"]}>
      {#if $$slots["tooltip-content"]}
        <TooltipContent><slot name="tooltip-content" /></TooltipContent>
      {/if}
    </div>
  </Tooltip>

  {#if showContextMenu}
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          builders={[builder]}
          id="more-actions-{name}"
          tooltipText="More actions"
          suppressTooltip={contextMenuOpen}
          label="{name} actions menu trigger"
        >
          <MoreHorizontal />
        </ContextButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        class="border-none bg-gray-800 text-white min-w-60"
        align="start"
        side="right"
        sideOffset={16}
      >
        <slot name="menu-items" />
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</li>

{#if showDetails}
  <slot name="more" />
{/if}

<style lang="postcss">
  .entry {
    @apply w-full justify-between pl-2.5 pr-2;
    @apply flex items-center h-6 select-none cursor-pointer;
  }

  .entry:focus {
    @apply outline-none;
  }

  .entry:hover:not(.open) {
    @apply bg-gray-100;
  }

  .entry:focus,
  .open {
    @apply bg-gray-200 font-bold;
  }

  .clickable-text {
    @apply text-left size-full overflow-hidden pl-6 flex items-center;
    @apply ui-copy gap-x-1 text-ellipsis;
  }

  .expandable {
    @apply pl-0;
  }
</style>
