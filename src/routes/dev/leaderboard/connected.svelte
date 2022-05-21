<script lang="ts">
  import { browser } from "$app/env";
  import { fly, fade } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { quadInOut as flipEasing, cubicIn } from "svelte/easing";

  import { tweened } from "svelte/motion";

  import Close from "$lib/components/icons/Close.svelte";
  /** for now, this LeaderboardFeature.svelte file will be here. */
  import Leaderboard from "./_LeaderboardFeature.svelte";

  import { getContext, onMount } from "svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    formatInteger,
    formatBigNumberPercentage,
  } from "$lib/util/formatters";

  import BarAndLabel from "$lib/components/BarAndLabel.svelte";

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

  /** scratch work */
</script>

<svelte:window on:resize={onResize} />
<div class="w-screen min-h-screen bg-white p-8">
  {#if $persistentTableStore?.entities}
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
  {/if}

  <section>
    <header
      style:grid-template-columns="max-content max-content"
      class="pb-6 pt-6 grid  w-full justify-between"
    >
      <h1 style:line-height="1.1">
        <div class="pl-2 text-gray-600 font-normal" style:font-size="1.5rem">
          Total Records
        </div>
        <div style:font-size="2rem" style:width="600px">
          <div class="w-full">
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

      <div>
        {#if anythingSelected}
          <!-- FIXME: we should be generalizing whatever this button is -->
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
    <div bind:this={leaderboardContainer}>
      <div
        style:grid-template-columns="repeat({columns}, 315px)"
        style:max-height="80vh"
        class="
          border-t border-gray-200
          overflow-auto
            grid
            gap-6 justify-start"
      >
        {#each leaderboards as { displayName, values }, i (displayName)}
          {@const nullCount = 0}
          <div
            style:width="315px"
            transition:fade={{
              duration: 200,
            }}
            animate:flip={{
              duration: 200,
              easing: flipEasing,
            }}
            style:grid-column={1 + (i % columns)}
            style:grid-row={1 + Math.floor(i / columns)}
          >
            <Leaderboard
              seeMore={leaderboardExpanded === displayName}
              on:expand={() => {
                if (leaderboardExpanded === displayName) {
                  leaderboardExpanded = undefined;
                  setTimeout(() => {
                    waitForLeaderboardClearout = false;
                  }, 600);
                } else {
                  leaderboardExpanded = displayName;
                  waitForLeaderboardClearout = true;
                }
              }}
              on:select-item={(event) => {
                activeValues[displayName];
                if (!activeValues[displayName].includes(event.detail)) {
                  activeValues[displayName] = [
                    ...activeValues[displayName],
                    event.detail,
                  ];
                } else {
                  activeValues[displayName] = activeValues[displayName]?.filter(
                    (b) => b !== event.detail
                  );
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
                activeValues[displayName] = [];
              }}
              activeValues={activeValues[displayName]}
              {displayName}
              {values}
              referenceValue={referenceValue || 0}
            />
          </div>
        {/each}
      </div>
    </div>
  </section>
</div>
