<script lang="ts">
  import { browser } from "$app/env";
  import { fly, fade } from "svelte/transition";
  import { quadInOut as flipEasing, cubicIn } from "svelte/easing";
  import { flip } from "svelte/animate";
  import VirtualizedGrid from "./_VirtualizedGrid.svelte";

  import { tweened } from "svelte/motion";

  import Close from "$lib/components/icons/Close.svelte";
  /** for now, this LeaderboardFeature.svelte file will be here. */
  import Leaderboard from "./_LeaderboardFeature.svelte";

  import CheckerFull from "$lib/components/icons/CheckerFull.svelte";
  import CheckerHalf from "$lib/components/icons/CheckerHalf.svelte";

  import FilterSet from "./_components/FilterSet.svelte";
  import Filter from "./_components/Filter.svelte";

  import { getContext, onMount } from "svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    formatInteger,
    formatBigNumberPercentage,
  } from "$lib/util/formatters";

  import BarAndLabel from "$lib/components/BarAndLabel.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  let leaderboards = [];
  let currentTable: string;

  let store;
  let persistentTableStore;
  let derivedTableStore;

  let bigNumber;
  /** This is the reference value used to scale the bars.
   * WHen it's a count(*) metric, it's identical to the leaderboard.
   */
  let referenceValue;
  const bigNumberTween = tweened(0, {
    duration: 1000,
    delay: 200,
    easing: cubicIn,
  });
  $: bigNumberTween.set(bigNumber || 0);

  const metricFormatters = {
    simpleSummable: formatInteger,
  };

  if (browser) {
    store = getContext("rill:app:store");
    persistentTableStore = getContext("rill:app:persistent-table-store");
    derivedTableStore = getContext("rill:app:derived-table-store");
    store.socket.emit("explorer", "ok!!!");
  }

  /** these should move to their own file. */
  let availableDimensions = [];
  if (browser) {
    /** listen to the available columns here. */
    store.socket.on("getAvailableDimensions", ({ dimensions }) => {
      console.log("where are we?", currentTable);
      availableDimensions = dimensions;
      // now, uh, calculate all the dimension leaderboards.
      availableDimensions.forEach((dimensionName) => {
        store.socket.emit("getDimensionLeaderboard", {
          dimensionName,
          entityType: EntityType.Table,
          entityID: currentTable,
        });
      });
    });
    // receive getDimensionLeaderboard responses.
    store.socket.on("getDimensionLeaderboard", ({ dimensionName, values }) => {
      let exists = leaderboards.find(
        (leaderboard) => leaderboard?.displayName === dimensionName
      );
      if (exists) {
        exists.values = values;
        exists.displayName = dimensionName;
      }

      if (exists) leaderboards = [...leaderboards];
      else
        leaderboards = [
          ...leaderboards,
          { displayName: dimensionName, values },
        ];
      // add to the activeValues.
      if (!(dimensionName in activeValues)) {
        activeValues[dimensionName] = [];
      }
    });
    // receive bigNumber
    store.socket.on("getBigNumber", ({ metric, value, filters }) => {
      bigNumber = value;
      if (!isAnythingSelected(filters)) {
        referenceValue = value;
      }
    });
  }
  /** ------------------------------------ */

  /** prunes the actives list to the bare minimum needed for the API. */
  function prune(actives) {
    return Object.keys(actives)
      .filter((key) => {
        return activeValues[key].length;
      })
      .reduce((acc, v) => {
        acc[v] = activeValues[v].map((value) => [value, "include"]);
        return acc;
      }, {});
  }

  function isAnythingSelected(filters): boolean {
    if (!filters) return false;
    return Object.keys(filters).some((key) => {
      return filters[key]?.length;
    });
  }

  /**
   * get the current leaderboard element.
   */
  let activeValues = {};

  function initializeActiveValues(leaderboards) {
    if (!leaderboards && !leaderboards.length) return {};
    return leaderboards.reduce((acc, leaderboard) => {
      acc[leaderboard.displayName] = [];
      return acc;
    }, {});
  }

  function clearAllFilters() {
    activeValues = initializeActiveValues(leaderboards);
    bigNumber = 0;
    store.socket.emit("getBigNumber", {
      entityType: EntityType.Table,
      entityID: currentTable,
      expression: "count(*)",
    });
    availableDimensions.forEach((dimensionName) => {
      store.socket.emit("getDimensionLeaderboard", {
        dimensionName,
        entityType: EntityType.Table,
        entityID: currentTable,
      });
    });
  }

  $: anythingSelected = isAnythingSelected(activeValues);

  let columns = 3;
  let leaderboardContainer: HTMLElement;
  let availableWidth = 0;

  function onResize() {
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.floor(availableWidth / (315 + 20));
  }

  onMount(() => {
    // determine initial resize.
    onResize();
    leaderboards = [];
  });

  $: if (!currentTable && $persistentTableStore?.entities?.length) {
    currentTable = $persistentTableStore?.entities[0].id;

    activeValues = initializeActiveValues(leaderboards);
    store.socket.emit("getAvailableDimensions", {
      entityType: EntityType.Table,
      entityID: currentTable,
    });
    store.socket.emit("getBigNumber", {
      entityType: EntityType.Table,
      entityID: currentTable,
      expression: "count(*)",
    });
  }

  let leaderboardExpanded: string;
  let waitForLeaderboardClearout = false;

  /** State for the reference value toggle */
  let whichReferenceValue = "filtered";
  $: stagedReferenceValue =
    whichReferenceValue === "filtered" ? bigNumber : referenceValue;

  /** scratch work */
</script>

<svelte:window on:resize={onResize} />
<div class="w-screen min-h-screen bg-white p-8">
  <!-- {#if $persistentTableStore?.entities}
    <select
      on:change={(event) => {
        currentTable = event.target.value;
        // this is where we re-establish the table names?
        leaderboards = [];
        activeValues = initializeActiveValues(leaderboards);
        store.socket.emit("getAvailableDimensions", {
          entityType: EntityType.Table,
          entityID: currentTable,
        });
        store.socket.emit("getBigNumber", {
          entityType: EntityType.Table,
          entityID: currentTable,
          expression: "count(*)",
        });
      }}
    >
      {#each $persistentTableStore?.entities as entity}
        <option value={entity.id}>{entity.tableName}</option>
      {/each}
    </select>
  {/if} -->

  <section>
    <header
      style:grid-template-columns="max-content auto"
      class="pb-6 pt-6 grid w-full justify-between"
    >
      <h1 style:line-height="1.1">
        <div class="pl-2 text-gray-600 font-normal" style:font-size="1.5rem">
          Total Records
        </div>
        <div style:font-size="2rem" style:width="400px">
          <div class="w-full rounded">
            <BarAndLabel
              justify="stretch"
              showBackground={anythingSelected}
              color={!anythingSelected ? "bg-transparent" : "bg-blue-200"}
              value={bigNumber / referenceValue || 0}
            >
              <div
                style:grid-template-columns="auto auto"
                class="grid items-center gap-x-2 w-full text-left pb-2 pt-2"
              >
                <div>
                  {metricFormatters.simpleSummable(~~$bigNumberTween)}
                </div>

                <div class="font-normal text-gray-600 italic text-right">
                  {#if $bigNumberTween && referenceValue}
                    {formatBigNumberPercentage(
                      $bigNumberTween / referenceValue
                    )}
                  {/if}
                </div>
              </div>
            </BarAndLabel>
          </div>
        </div>
      </h1>
      <!-- 
      <div
        style:font-size="24px"
        class="grid justify-items-start justify-start grid-flow-col items-center"
      >
        <Tooltip distance={16}>
          <button
            class="m-0 p-1 transition-color"
            class:bg-transparent={whichReferenceValue !== "filtered"}
            class:bg-gray-200={whichReferenceValue === "filtered"}
            class:font-bold={whichReferenceValue === "filtered"}
            class:text-gray-400={whichReferenceValue !== "filtered"}
            on:click={() => (whichReferenceValue = "filtered")}
            ><CheckerHalf /></button
          >
          <TooltipContent slot="tooltip-content">
            scale leaderboard bars by currently-filtered total
          </TooltipContent>
        </Tooltip>
        <Tooltip distance={16}>
          <button
            class="m-0 p-1 transition-color"
            class:bg-transparent={whichReferenceValue !== "global"}
            class:bg-gray-200={whichReferenceValue === "global"}
            class:font-bold={whichReferenceValue === "global"}
            class:text-gray-400={whichReferenceValue !== "global"}
            on:click={() => (whichReferenceValue = "global")}
            ><CheckerFull /></button
          >
          <TooltipContent slot="tooltip-content">
            scale leaderboard bars by total record count
          </TooltipContent>
        </Tooltip>
      </div> -->

      <div>
        {#if anythingSelected}
          <!-- FIXME: we should be generalizing whatever this button is -->
          <div class="flex flex-col gap-y-1">
            {#each Object.keys(activeValues) as dimension, i}
              {#if activeValues[dimension].length}
                <FilterSet>
                  <div transition:fly={{ duration: 200, x: -16 }} slot="name">
                    {dimension}
                  </div>
                  <svelte:fragment slot="values">
                    {#each activeValues[dimension] as value (dimension + value)}
                      <div
                        animate:flip={{ duration: 200 }}
                        transition:fly={{ duration: 200, x: 16 }}
                      >
                        <Filter
                          on:click={() => {
                            activeValues[dimension] = activeValues[
                              dimension
                            ]?.filter((b) => b !== value);
                            if (browser) {
                              const filters = prune(activeValues);
                              bigNumber = 0;

                              store.socket.emit("getBigNumber", {
                                entityType: EntityType.Table,
                                entityID: currentTable,
                                expression: "count(*)",
                                filters,
                              });
                              availableDimensions.forEach((dimensionName) => {
                                // invalidate the exiting leaderboard?
                                store.socket.emit("getDimensionLeaderboard", {
                                  dimensionName,
                                  entityType: EntityType.Table,
                                  entityID: currentTable,
                                  filters,
                                });
                              });
                            }
                          }}
                        >
                          {value}
                        </Filter>
                      </div>
                    {/each}
                  </svelte:fragment>
                </FilterSet>
              {/if}
            {/each}
          </div>

          <button
            transition:fly={{ duration: 200, y: 5 }}
            on:click={clearAllFilters}
            class="
                        grid gap-x-2 items-center font-bold
                        bg-red-100
                        text-red-900
                        p-1
                        pl-2 pr-2
                        rounded
                    "
            style:grid-template-columns="auto max-content"
          >
            clear all filters <Close />
          </button>
        {/if}
      </div>
    </header>
    <div>
      <div
        bind:this={leaderboardContainer}
        style:height="80vh"
        class="
          border-t border-gray-200
          "
      >
        <VirtualizedGrid {columns} height="100%" items={leaderboards} let:item>
          <!-- the single virtual element -->
          <div style:width="315px">
            <Leaderboard
              seeMore={leaderboardExpanded === item.displayName}
              on:expand={() => {
                if (leaderboardExpanded === item.displayName) {
                  leaderboardExpanded = undefined;
                  setTimeout(() => {
                    waitForLeaderboardClearout = false;
                  }, 600);
                } else {
                  leaderboardExpanded = item.displayName;
                  waitForLeaderboardClearout = true;
                }
              }}
              on:select-item={(event) => {
                activeValues[item.displayName];
                if (!activeValues[item.displayName].includes(event.detail)) {
                  activeValues[item.displayName] = [
                    ...activeValues[item.displayName],
                    event.detail,
                  ];
                } else {
                  activeValues[item.displayName] = activeValues[
                    item.displayName
                  ]?.filter((b) => b !== event.detail);
                }

                if (browser) {
                  const filters = prune(activeValues);
                  bigNumber = 0;

                  store.socket.emit("getBigNumber", {
                    entityType: EntityType.Table,
                    entityID: currentTable,
                    expression: "count(*)",
                    filters,
                  });
                  availableDimensions.forEach((dimensionName) => {
                    // invalidate the exiting leaderboard?
                    store.socket.emit("getDimensionLeaderboard", {
                      dimensionName,
                      entityType: EntityType.Table,
                      entityID: currentTable,
                      filters,
                    });
                  });
                }
              }}
              on:clear-all={() => {
                activeValues[item.displayName] = [];
              }}
              activeValues={activeValues[item.displayName]}
              displayName={item.displayName}
              values={item.values}
              referenceValue={stagedReferenceValue || 0}
            />
          </div>
          <!-- {/each} -->
        </VirtualizedGrid>
      </div>
    </div>
  </section>
</div>
