<script lang="ts">
  import { SvelteComponent, tick } from "svelte/internal";
  import { onMount, createEventDispatcher } from "svelte";
  // import { slide } from "svelte/transition";
  import { tweened } from "svelte/motion";
  import { cubicInOut as easing, cubicOut } from "svelte/easing";
  import { format } from "d3-format";

  import type { EntityId } from "@reduxjs/toolkit";

  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  // import * as classes from "$lib/util/component-classes";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  // import ColumnProfile from "./ColumnProfile.svelte";

  import Spacer from "$lib/components/icons/Spacer.svelte";

  import NavEntry from "$lib/components/column-profile/NavEntry.svelte";

  import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";

  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";

  import {
    defaultSort,
    // sortByNullity,
    // sortByName,
  } from "$lib/components/column-profile/sort-utils";
  import notificationStore from "$lib/components/notifications/";

  import { onClickOutside } from "$lib/util/on-click-outside";
  // import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import { store, reduxReadable } from "$lib/redux-store/store-root";

  import { deleteMetricsDef } from "$lib/redux-store/metrics-definition/metrics-definition-slice";

  export let metricsDefId: EntityId;

  export let name: string;
  // export let cardinality: number;
  // export let profile: any;
  // export let head: any; // FIXME
  // export let sizeInBytes: number = undefined;
  export let emphasizeTitle = false;
  // export let draggable = true;
  export let show = false;
  // export let showTitle = true;
  // export let showContextButton = true;
  // export let indentLevel = 0;

  $: metricsDef = $reduxReadable?.metricsDefinition?.entities[metricsDefId];
  $: name = metricsDef?.metricDefLabel;

  const dispatch = createEventDispatcher();

  // const formatInteger = format(",");

  // let containerWidth = 0;
  let contextMenu;
  let contextMenuOpen = false;
  // let container;

  // onMount(() => {
  //   const observer = new ResizeObserver((entries) => {
  //     containerWidth = container?.clientWidth ?? 0;
  //   });
  //   observer.observe(container);
  // });

  // let cardinalityTween = tweened(cardinality, { duration: 600, easing });
  // let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });

  // $: cardinalityTween.set(cardinality || 0);
  // $: interimCardinality = ~~$cardinalityTween;
  // $: sizeTween.set(sizeInBytes || 0);

  // let selectingColumns = false;
  // let selectedColumns = [];

  // let sortedProfile;
  // const sortByOriginalOrder = null;

  // let sortMethod = defaultSort;
  // $: if (sortMethod !== sortByOriginalOrder) {
  //   sortedProfile = [...profile].sort(sortMethod);
  // } else {
  //   sortedProfile = profile;
  // }

  // let previewView = "summaries";

  let menuX;
  let menuY;
  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

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
    dispatch("select");
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
          <!-- <span
            ><span
              >{cardinality !== undefined && !isNaN(cardinality)
                ? formatInteger(interimCardinality)
                : "no"}</span
            >
            row{#if cardinality !== 1}s{/if}</span
          > -->
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
      <!-- {/if} -->
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
        <Menu
          on:escape={() => {
            console.log("esc");
            contextMenuOpen = false;
          }}
          on:item-select={() => {
            console.log("item-select");
            contextMenuOpen = false;
          }}
        >
          <MenuItem
            on:select={() => {
              console.log("select");
              store.dispatch(deleteMetricsDef(metricsDefId));
            }}
          >
            delete {name}
          </MenuItem>
        </Menu>
      </FloatingElement>
    </div>
  {/if}
</NavEntry>
