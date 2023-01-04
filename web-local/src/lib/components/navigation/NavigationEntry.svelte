<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { createCommandClickAction } from "../../util/command-click-action";
  import ContextButton from "../column-profile/ContextButton.svelte";
  import ExpanderButton from "../column-profile/ExpanderButton.svelte";

  export let name: string;
  export let href: string;
  export let open = false;
  export let notExpandable = false;
  export let tooltipMaxWidth: string = undefined;

  const { commandClickAction } = createCommandClickAction();
  const { shiftClickAction } = createShiftClickAction();

  let showDetails = false;
  let contextMenuOpen = false;

  function onShowDetails() {
    showDetails = !showDetails;
  }

  const shiftClickHandler = async () => {
    await navigator.clipboard.writeText(name);
    notifications.send({ message: `copied "${name}" to clipboard` });
  };

  let containerFocused;
  let contextMenuHovered = false;
  let seeMoreHovered = false;

  function onContainerFocus(tf) {
    return () => {
      containerFocused = tf;
    };
  }
</script>

<Tooltip
  location="right"
  alignment="start"
  distance={16}
  suppress={contextMenuHovered || contextMenuOpen || seeMoreHovered}
>
  <div
    on:mouseenter={onContainerFocus(true)}
    on:focus={onContainerFocus(true)}
    on:mouseleave={onContainerFocus(false)}
    on:blur={onContainerFocus(false)}
    style:height="24px"
    class:font-bold={open}
    class:bg-gray-200={open}
    class="navigation-entry-title grid gap-x-1 items-center pl-4 pr-3 {!open
      ? 'hover:bg-gray-100'
      : ''}"
    style:grid-template-columns="max-content auto max-content"
    use:commandClickAction
    use:shiftClickAction
    on:command-click
    on:shift-click={shiftClickHandler}
  >
    <!-- slot for navigation click -->

    <div class="mr-1">
      {#if !notExpandable}
        <ExpanderButton
          bind:isHovered={seeMoreHovered}
          rotated={showDetails}
          on:click={onShowDetails}
        >
          <CaretDownIcon size="14px" />
        </ExpanderButton>
      {:else}
        <Spacer size="16px" />
      {/if}
    </div>

    <a
      class="ui-copy flex items-center gap-x-2 w-full text-ellipsis overflow-hidden whitespace-nowrap"
      {href}
      on:click={() => {
        if (open) onShowDetails();
      }}
    >
      {#if $$slots["icon"]}
        <div class="text-gray-400" style:width="1em" style:height="1em">
          <slot name="icon" />
        </div>
      {/if}
      <div class=" text-ellipsis overflow-hidden whitespace-nowrap">
        {name}
      </div>
    </a>

    <!-- context menu -->
    <WithTogglableFloatingElement
      location="right"
      alignment="start"
      distance={16}
      let:toggleFloatingElement
      bind:active={contextMenuOpen}
    >
      <span
        class="self-center"
        class:opacity-0={!containerFocused &&
          !contextMenuOpen &&
          !open &&
          !contextMenuHovered}
      >
        <ContextButton
          id="more-actions-{name}"
          tooltipText="More actions"
          suppressTooltip={contextMenuOpen}
          on:click={toggleFloatingElement}
          bind:isHovered={contextMenuHovered}
          width={24}
          height={24}
          border={false}
        >
          <MoreHorizontal />
        </ContextButton>
      </span>
      <Menu
        dark
        on:click-outside={toggleFloatingElement}
        on:escape={toggleFloatingElement}
        on:item-select={toggleFloatingElement}
        slot="floating-element"
      >
        <slot name="menu-items" toggleMenu={toggleFloatingElement} />
      </Menu>
    </WithTogglableFloatingElement>
  </div>
  <!-- if tooltip content is present in a slot, render the tooltip -->
  <div slot="tooltip-content" class:hidden={!$$slots["tooltip-content"]}>
    {#if $$slots["tooltip-content"]}
      <TooltipContent maxWidth={tooltipMaxWidth}
        ><slot name="tooltip-content" /></TooltipContent
      >
    {/if}
  </div>
</Tooltip>

{#if showDetails}
  <slot name="more" />
{/if}
