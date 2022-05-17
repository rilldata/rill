<script lang="ts">
  import { fly } from "svelte/transition";
  import Close from "$lib/components/icons/Close.svelte";
  /** for now, this LeaderboardFeature.svelte file will be here. */
  import Leaderboard from "./_LeaderboardFeature.svelte";

  import { swimLanePlacement } from "$lib/util/swim-lane-placement";
  import { onMount } from "svelte";

  /** remove this before we componentize anything. */
  let files = [];
  try {
    // @ts-ignore
    files = import.meta.globEager("./data/*.json");
  } catch (err) {
    console.log("initial build did not work out.");
  }

  const leaderboardSet = Object.keys(files).map((fileName) => {
    const leaderboard = files[fileName].default;
    return [fileName, leaderboard];
  });

  /**
   * get the current leaderboard element.
   */
  let currentLeaderboard = leaderboardSet.length ? leaderboardSet[0][1] : [];
  let activeValues = {};

  function initializeActiveValues(leaderboards) {
    if (!leaderboards) return [];
    return leaderboards.reduce((acc, leaderboard) => {
      acc[leaderboard.displayName] = [];
      return acc;
    }, {});
  }

  function clearAllFilters() {
    activeValues = initializeActiveValues(currentLeaderboard?.leaderboards);
  }

  $: activeValues = initializeActiveValues(currentLeaderboard?.leaderboards);

  $: anythingSelected = Object.keys(activeValues).some((key) => {
    return activeValues[key]?.length;
  });

  let columns = 3;
  let leaderboardContainer: Element;
  let availableWidth = 0;
  function onResize() {
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.floor(availableWidth / (315 + 20));
  }

  onMount(() => {
    // determine initial resize.
    onResize();
  });
</script>

<svelte:window on:resize={onResize} />
repeat({columns}, max-content)
<div class="w-screen min-h-screen bg-white p-8">
  <section>
    {#if leaderboardSet.length}
      <header
        style:height="32px"
        style:grid-template-columns="max-content max-content"
        class="pb-3 grid  w-full justify-between"
      >
        <select bind:value={currentLeaderboard}>
          {#each leaderboardSet as [file, leaderboard]}
            <option value={leaderboard}>{leaderboard.displayName}</option>
          {/each}
        </select>
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
          style:grid-template-columns="repeat({columns}, max-content)"
          class="
            grid 
            gap-6 justify-start w-max"
        >
          <!-- {#each currentLeaderboard.leaderboards as {displayName, values, nullCount }} -->
          {#each swimLanePlacement(currentLeaderboard.leaderboards, (leaderboard) => leaderboard.values.length, columns) as lane, i}
            <div class="flex flex-col">
              {#each lane as { displayName, values, nullCount }, j (displayName)}
                <Leaderboard
                  on:select-item={(event) => {
                    activeValues[displayName];
                    if (!activeValues[displayName].includes(event.detail)) {
                      activeValues[displayName] = [
                        ...activeValues[displayName],
                        event.detail,
                      ];
                    } else {
                      activeValues[displayName] = activeValues[
                        displayName
                      ]?.filter((b) => b !== event.detail);
                    }
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
              {/each}
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <p>
        <b>No leaderboards present.</b>
      </p>
      <p style:width="600px">
        Run <code class="italic text-blue-600"
          >node ./scripts/dev/generate-leaderboards.js path/to/stage.db</code
        >
        to generate example leaderboard data.
      </p>
    {/if}
  </section>
</div>
