<script lang="ts">
  import { goto } from "$app/navigation";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import ContextButton from "@rilldata/web-local/lib/components/column-profile/ContextButton.svelte";
  import ExpanderButton from "@rilldata/web-local/lib/components/column-profile/ExpanderButton.svelte";
  import { createCommandClickAction } from "@rilldata/web-local/lib/util/command-click-action";
  import { currentHref } from "./stores";

  export let name: string;
  export let href: string;
  export let open = false;
  export let expandable = true;
  export let tooltipMaxWidth: string = undefined;
  export let maxMenuWidth: string = undefined;

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

  /** track the mousedown event to provide immediate feedback on the click motion. */
  let mousedown = false;
  /** We'll look at the first two segments of the pathname in order to determine
   * if this entry is active. So it should only need to match e.g. /dashboard/<name>,
   * which maintains any subroutes beyond that (like /dashboard/<name>/edit).
   */
  $: pathname = $currentHref.split("/").slice(0, 3).join("/");
  $: isActive = pathname === href;
</script>

<Tooltip
  location="right"
  alignment="start"
  distance={16}
  suppress={contextMenuHovered || contextMenuOpen || seeMoreHovered}
>
  <a
    {href}
    on:click={() => {
      if (open) onShowDetails();
    }}
    on:mousedown={() => {
      currentHref.set(href);
      mousedown = true;
    }}
    on:mouseup={() => {
      mousedown = false;
    }}
    on:mouseenter={onContainerFocus(true)}
    on:focus={onContainerFocus(true)}
    on:mouseleave={onContainerFocus(false)}
    on:blur={onContainerFocus(false)}
    on:dragend={async () => {
      // perform navigation in this case.
      await goto(href);
      mousedown = false;
      currentHref.set(href);
    }}
    style:height="24px"
    class:font-bold={isActive}
    class:bg-gray-200={isActive}
    class:bg-gray-100={isActive && mousedown}
    class="
    navigation-entry-title grid gap-x-1 items-center pl-2 pr-2 {!isActive &&
    !mousedown
      ? 'hover:bg-gray-100'
      : ''}"
    style:grid-template-columns="max-content auto max-content"
    use:commandClickAction
    use:shiftClickAction
    on:command-click
    on:shift-click={shiftClickHandler}
  >
    <!-- slot for navigation click -->

    <div>
      {#if expandable}
        <ExpanderButton
          bind:isHovered={seeMoreHovered}
          rotated={showDetails}
          on:click={onShowDetails}
        >
          <CaretDownIcon size="14px" />
        </ExpanderButton>
      {:else}
        <Spacer size="14px" />
      {/if}
    </div>

    <div
      class="ui-copy flex items-center gap-x-1 w-full text-ellipsis overflow-hidden whitespace-nowrap"
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
    </div>

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
          !isActive &&
          !contextMenuHovered}
      >
        <ContextButton
          id="more-actions-{name}"
          tooltipText="More actions"
          suppressTooltip={contextMenuOpen}
          on:click={(event) => {
            /** prevent the link click from registering */
            event.stopPropagation();
            toggleFloatingElement();
          }}
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
        maxWidth={maxMenuWidth}
        on:click-outside={toggleFloatingElement}
        on:escape={toggleFloatingElement}
        on:item-select={toggleFloatingElement}
        slot="floating-element"
      >
        <slot name="menu-items" toggleMenu={toggleFloatingElement} />
      </Menu>
    </WithTogglableFloatingElement>
  </a>
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
