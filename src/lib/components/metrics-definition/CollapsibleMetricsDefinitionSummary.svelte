<script lang="ts">
  import { SvelteComponent, tick } from "svelte/internal";
  import { onMount, createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { tweened } from "svelte/motion";
  import { cubicInOut as easing, cubicOut } from "svelte/easing";
  import { format } from "d3-format";
  import type { EntityId } from "@reduxjs/toolkit";

  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import * as classes from "$lib/util/component-classes";
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

  import CollapsibleMetricsDefinitionSummaryNavEntry from "./CollapsibleMetricsDefinitionSummaryNavEntry.svelte";

  import {
    defaultSort,
    sortByNullity,
    sortByName,
  } from "$lib/components/column-profile/sort-utils";
  import notificationStore from "$lib/components/notifications/";

  import { onClickOutside } from "$lib/util/on-click-outside";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";

  export let metricsDefId: EntityId;
  export let name: string;
  export let cardinality: number;
  // export let profile: any;
  export let head: any; // FIXME
  export let sizeInBytes: number = undefined;
  export let emphasizeTitle = false;
  export let draggable = true;
  export let show = false;
  export let showTitle = true;
  export let showContextButton = true;
  export let indentLevel = 0;

  const dispatch = createEventDispatcher();

  const formatInteger = format(",");

  let containerWidth = 0;
  let contextMenu;
  let contextMenuOpen = false;
  let container;

  onMount(() => {
    const observer = new ResizeObserver((entries) => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
  });

  let cardinalityTween = tweened(cardinality, { duration: 600, easing });
  let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });

  $: cardinalityTween.set(cardinality || 0);
  $: interimCardinality = ~~$cardinalityTween;
  $: sizeTween.set(sizeInBytes || 0);

  let selectingColumns = false;
  let selectedColumns = [];

  // let sortedProfile;
  const sortByOriginalOrder = null;

  let sortMethod = defaultSort;
  // $: if (sortMethod !== sortByOriginalOrder) {
  //   sortedProfile = [...profile].sort(sortMethod);
  // } else {
  //   sortedProfile = profile;
  // }

  let previewView = "summaries";

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

<div bind:this={container}>
  {#if showTitle}
    <div {draggable} class="active:cursor-grabbing">
      <CollapsibleMetricsDefinitionSummaryNavEntry {metricsDefId} />
    </div>
  {/if}
  <!-- 
  {#if show}
    <div
      class="pt-1 pb-3 pl-accordion"
      transition:slide|local={{ duration: 120 }}
    >
      <div
        class="pl-{indentLevel === 1
          ? '10'
          : '4'} pr-5 pb-2 flex justify-between text-gray-500"
        class:flex-col={containerWidth < 325}
      >
        <select
          style:transform="translateX(-4px)"
          bind:value={sortMethod}
          class={classes.NATIVE_SELECT}
        >
          <option value={sortByOriginalOrder}>show original order</option>
          <option value={defaultSort}>sort by type</option>
          <option value={sortByNullity}>sort by null %</option>
          <option value={sortByName}>sort by name</option>
        </select>
        <select
          style:transform="translateX(4px)"
          bind:value={previewView}
          class={classes.NATIVE_SELECT}
          class:hidden={containerWidth < 325}
        >
          <option value="summaries">show summary&nbsp;</option>
          <option value="example">show example</option>
        </select>
      </div>

      <div>
        {#if sortedProfile && sortedProfile.length && head.length}
          {#each sortedProfile as column (column.name)}
             <ColumnProfile
              {indentLevel}
              example={head[0][column.name] || ""}
              {containerWidth}
              hideNullPercentage={containerWidth <
                COLUMN_PROFILE_CONFIG.hideNullPercentage}
              hideRight={containerWidth < COLUMN_PROFILE_CONFIG.hideRight}
              compactBreakpoint={COLUMN_PROFILE_CONFIG.compactBreakpoint}
              view={previewView}
              name={column.name}
              type={column.type}
              summary={column.summary}
              totalRows={cardinality}
              nullCount={column.nullCount}
            >
              <button slot="context-button" class:hidden={!showContextButton}>
                <Spacer size="16px" />
              </button>
            </ColumnProfile>
          {/each}
        {/if}
      </div>
    </div>
  {/if} -->
</div>
