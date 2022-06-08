<script lang="ts">
  import { SvelteComponent, tick } from "svelte/internal";
  import { onMount, createEventDispatcher } from "svelte";
  import type { EntityId } from "@reduxjs/toolkit";

  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  // import * as classes from "$lib/util/component-classes";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  // import ColumnProfile from "./ColumnProfile.svelte";

  import NavEntry from "$lib/components/column-profile/NavEntry.svelte";

  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";

  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";

  import notificationStore from "$lib/components/notifications/";

  import { onClickOutside } from "$lib/util/on-click-outside";
  import { store, reduxReadable } from "$lib/redux-store/store-root";

  import { deleteMetricsDef } from "$lib/redux-store/metrics-definition/metrics-definition-slice";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  // FIXME: wnat to remove svelte stores and contexts
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { getContext } from "svelte";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  export let metricsDefId: EntityId;

  export let name: string;
  // export let emphasizeTitle = false;
  export let show = false;

  $: metricsDef = $reduxReadable?.metricsDefinition?.entities[metricsDefId];
  $: name = metricsDef?.metricDefLabel;

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;
  $: emphasizeTitle = $rillAppStore?.activeEntity?.id === metricsDefId;

  const dispatch = createEventDispatcher();

  let contextMenu;
  let contextMenuOpen = false;
  const closeContextMenu = () => {
    contextMenuOpen = false;
  };

  let menuX;
  let menuY;
  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  const dispatchDeleteMetricsDef = () => {
    store.dispatch(deleteMetricsDef(metricsDefId));
  };

  // state for title bar hover.
  let titleElementHovered = false;
</script>

<NavEntry
  expanded={show}
  selected={emphasizeTitle}
  bind:hovered={titleElementHovered}
  on:shift-click={async () => {
    await navigator.clipboard.writeText(name);
    notificationStore.send({ message: `copied "${name}" to clipboard` });
  }}
  on:select-body={async (event) => {
    dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsDef,
      metricsDefId.toString(), // FIXME: should not need to do this type conversion
    ]);
    // dispatch("select");
  }}
  on:expand={() => {
    show = !show;
    // pass up expand
    dispatch("expand");
  }}
>
  <svelte:fragment slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        {name}
      </svelte:fragment>
      <svelte:fragment slot="description" />
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>open in workspace</div>
      <Shortcut>click</Shortcut>
      <div>
        <StackingWord>copy</StackingWord> to clipboard
      </div>
      <Shortcut>shift + click</Shortcut>
    </TooltipShortcutContainer>
  </svelte:fragment>
  <!-- note: the classes in this span are also used for UI tests. -->
  <span
    class="collapsible-table-summary-title w-full"
    class:is-active={emphasizeTitle}
    class:font-bold={emphasizeTitle}
    class:italic={false}
  >
    {name}
  </span>
  <svelte:fragment slot="contextual-information">
    <div class="italic text-gray-600">
      <span
        class="grid grid-flow-col gap-x-2 text-gray-500 text-clip overflow-hidden whitespace-nowrap "
      >
        {#if titleElementHovered || emphasizeTitle}
          <span class="self-center">
            <ContextButton
              id={metricsDefId.toString()}
              tooltipText="delete"
              suppressTooltip={contextMenuOpen}
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
              }}><MoreIcon /></ContextButton
            >
          </span>
        {/if}
      </span>
    </div>
  </svelte:fragment>

  {#if contextMenuOpen}
    <div bind:this={contextMenu}>
      <FloatingElement
        relationship="mouse"
        target={{ x: menuX, y: menuY }}
        location="right"
        alignment="start"
      >
        <Menu on:escape={closeContextMenu} on:item-select={closeContextMenu}>
          <MenuItem on:select={dispatchDeleteMetricsDef}>
            delete {name}
          </MenuItem>
        </Menu>
      </FloatingElement>
    </div>
  {/if}
</NavEntry>
