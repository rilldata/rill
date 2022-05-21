<script lang="ts">
  import { browser } from "$app/env";
  import { fly, fade } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { quadInOut as flipEasing } from "svelte/easing";

  import Close from "$lib/components/icons/Close.svelte";
  /** for now, this LeaderboardFeature.svelte file will be here. */
  import Leaderboard from "./_LeaderboardFeature.svelte";

  import { getContext, onMount } from "svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  let leaderboards = [];
  let currentTable: string;

  let store;
  let persistentTableStore;
  let derivedTableStore;

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
      availableDimensions = dimensions;
      // now, uh, calculate all the dimension leaderboards.
      store.socket.emit("getDimensionLeaderboard", {
        dimensionName: availableDimensions[0],
        entityType: EntityType.Table,
        entityID: currentTable,
      });
      // availableDimensions.forEach((dimensionName) => {
      //   store.socket.emit("");
      // });
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
        acc[v] = activeValues[v].map((value) => ({ include: value }));
        return acc;
      }, {});
  }

  /**
   * get the current leaderboard element.
   */
  let activeValues = {};

  function initializeActiveValues(leaderboards) {
    if (!leaderboards) return [];
    return leaderboards.reduce((acc, leaderboard) => {
      acc[leaderboard.displayName] = [];
      return acc;
    }, {});
  }

  function clearAllFilters() {
    activeValues = initializeActiveValues(leaderboards);
  }

  $: activeValues = initializeActiveValues(leaderboards);

  $: anythingSelected = Object.keys(activeValues).some((key) => {
    return activeValues[key]?.length;
  });

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
  });

  let leaderboardExpanded: string;
  let waitForLeaderboardClearout = false;

  /** scratch work */
  $: console.log($persistentTableStore?.entities);
</script>

<svelte:window on:resize={onResize} />
<div class="w-screen min-h-screen bg-white p-8">
  {#if $persistentTableStore?.entities}
    <select
      on:change={(event) => {
        console.log(event.target.value);
        currentTable = event.target.value;
        // this is where we re-establish the table names?
        store.socket.emit("getAvailableDimensions", {
          entityType: EntityType.Table,
          entityID: currentTable,
        });
      }}
    >
      {#each $persistentTableStore?.entities as entity}
        <option value={entity.id}>{entity.tableName}</option>
      {/each}
    </select>
  {/if}
  {#each availableDimensions as dimension}
    <div>
      {dimension}
    </div>
  {/each}
  <section>
    <header
      style:height="32px"
      style:grid-template-columns="max-content max-content"
      class="pb-3 grid  w-full justify-between"
    >
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
        style:position="relative"
        style:grid-template-columns="repeat({columns}, 315px)"
        class="
            grid
            gap-6 justify-start"
      >
        {#each leaderboards as { displayName, values, nullCount }, i (displayName)}
          <div
            style:width="315px"
            transition:fade={{
              duration: 200,
              delay: waitForLeaderboardClearout ? 600 : 0,
            }}
            animate:flip={{
              duration:
                waitForLeaderboardClearout ||
                leaderboardExpanded === displayName
                  ? 600
                  : 200,
              easing: flipEasing,
              delay:
                waitForLeaderboardClearout &&
                leaderboardExpanded !== displayName
                  ? 200
                  : 0,
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

                if (browser) store.socket.emit("explorer", prune(activeValues));
              }}
              on:clear-all={() => {
                activeValues[displayName] = [];
              }}
              activeValues={activeValues[displayName]}
              {displayName}
              {values}
              total={currentLeaderboard.total}
              {nullCount}
            />
          </div>
        {/each}
      </div>
    </div>
  </section>
</div>
