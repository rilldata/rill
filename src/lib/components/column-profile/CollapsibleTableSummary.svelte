<script lang="ts">
  import { onMount, createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";

  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import * as classes from "$lib/util/component-classes";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

  import ColumnProfile from "./ColumnProfile.svelte";
  import CollapsibleTableHeader from "./CollapsibleTableHeader.svelte";

  import Spacer from "$lib/components/icons/Spacer.svelte";

  import {
    defaultSort,
    sortByNullity,
    sortByName,
  } from "$lib/components/column-profile/sort-utils";

  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  export let entityType: EntityType;
  export let name: string;
  export let cardinality: number;
  export let profile: any;
  export let head: any; // FIXME
  export let sizeInBytes: number = undefined;
  export let active = false;
  export let draggable = true;
  export let show = false;
  export let showTitle = true;
  export let showContextButton = true;
  export let indentLevel = 0;

  const dispatch = createEventDispatcher();

  let containerWidth = 0;
  let contextMenu;
  let contextMenuOpen;
  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });

  let sortedProfile;
  const sortByOriginalOrder = null;

  let sortMethod = defaultSort;
  $: if (sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
  } else {
    sortedProfile = profile;
  }

  let previewView = "summaries";

  let menuX;
  let menuY;
</script>

<div bind:this={container}>
  {#if showTitle}
    <div {draggable} class="active:cursor-grabbing">
      <CollapsibleTableHeader
        on:select
        on:query
        bind:contextMenuOpen
        bind:menuX
        bind:menuY
        bind:name
        bind:show
        {entityType}
        {contextMenu}
        {cardinality}
        {sizeInBytes}
        {active}
      />
    </div>
    {#if contextMenuOpen}
      <!-- place this above codemirror.-->
      <div bind:this={contextMenu}>
        <FloatingElement
          relationship="mouse"
          target={{ x: menuX, y: menuY }}
          location="right"
          alignment="start"
        >
          <Menu
            on:escape={() => {
              contextMenuOpen = false;
            }}
            on:item-select={() => {
              contextMenuOpen = false;
            }}
          >
            {#if entityType == EntityType.Table}
              <MenuItem
                on:select={() => {
                  dispatch("query");
                }}
              >
                query {name}
              </MenuItem>
            {/if}
            <MenuItem
              on:select={() => {
                dispatch("rename");
              }}
            >
              rename {name}
            </MenuItem>
            <MenuItem
              on:select={() => {
                dispatch("delete");
              }}
            >
              delete {name}
            </MenuItem>
          </Menu>
        </FloatingElement>
      </div>
    {/if}
  {/if}

  {#if show}
    <div
      class="pt-1 pb-3 pl-accordion"
      transition:slide|local={{ duration: 120 }}
    >
      <!-- pl-16 -->
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
  {/if}
</div>
