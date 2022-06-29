<script lang="ts">
  import { tick } from "svelte/internal";

  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import NavEntry from "$lib/components/column-profile/NavEntry.svelte";
  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import notificationStore from "$lib/components/notifications/";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { store } from "$lib/redux-store/store-root";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { getContext } from "svelte";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import ExpandCaret from "$lib/components/icons/ExpandCaret.svelte";
  import { toggleMetricsDefSummaryInNav } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { deleteMetricsDefsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";

  export let metricsDefId: string;

  $: thisMetricsDef = getMetricsDefReadableById(metricsDefId);

  $: name = $thisMetricsDef?.metricDefLabel;
  $: summaryExpanded = $thisMetricsDef?.summaryExpandedInNav;
  const rillAppStore = getContext("rill:app:store") as ApplicationStore;
  $: emphasizeTitle = $rillAppStore?.activeEntity?.id === metricsDefId;
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
  const contextButtonClickHandler = async (event) => {
    contextMenuOpen = !contextMenuOpen;
    menuX = event.clientX;
    menuY = event.clientY;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        contextMenuOpen = false;
      }, contextMenu);
    }
  };
  const dispatchDeleteMetricsDef = () => {
    store.dispatch(deleteMetricsDefsApi(metricsDefId));
  };
  const dispatchToggleSummaryInNav = () => {
    store.dispatch(toggleMetricsDefSummaryInNav(metricsDefId));
  };
  // state for title bar hover.
  let titleElementHovered = false;
</script>

<NavEntry
  expanded={summaryExpanded}
  selected={emphasizeTitle}
  bind:hovered={titleElementHovered}
  on:shift-click={async () => {
    await navigator.clipboard.writeText(name);
    notificationStore.send({ message: `copied "${name}" to clipboard` });
  }}
  on:select-body={async (event) => {
    dataModelerService.dispatch("setActiveAsset", [
      EntityType.MetricsDefinition,
      metricsDefId,
    ]);
  }}
  on:expand={dispatchToggleSummaryInNav}
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
              id={metricsDefId}
              tooltipText="delete"
              suppressTooltip={contextMenuOpen}
              on:click={contextButtonClickHandler}><MoreIcon /></ContextButton
            >
          </span>
          <span class="self-center">
            <ContextButton
              id={metricsDefId}
              tooltipText="expand"
              on:click={() => {
                dataModelerService.dispatch("setActiveAsset", [
                  EntityType.MetricsLeaderboard,
                  metricsDefId,
                ]);
              }}><ExpandCaret /></ContextButton
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
