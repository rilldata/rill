<script lang="ts">
  import { fly, fade } from "svelte/transition";
  import { flip } from "svelte/animate";
  //import { createFlipAnimationFactory } from "./_custom-flip";
  import { quadInOut as flipEasing } from "svelte/easing";
  import Close from "$lib/components/icons/Close.svelte";
  /** for now, this LeaderboardFeature.svelte file will be here. */
  import Leaderboard from "./_LeaderboardFeature.svelte";

  import { swimLanePlacement } from "$lib/util/swim-lane-placement";
  import { onMount } from "svelte";

  // const { flip, isFlipped } = createFlipAnimationFactory();

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
  let currentLeaderboard = leaderboardSet.length ? leaderboardSet[4][1] : [];
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
  function handleLeaderboarding(leaderboards, expandedLeaderboard = undefined) {
    return expandedLeaderboard
      ? leaderboards.filter((l) => l.displayName === expandedLeaderboard)
      : leaderboards;
  }
</script>

<svelte:window on:resize={onResize} />
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
          style:position="relative"
          style:grid-template-columns="repeat({leaderboardExpanded
            ? 1
            : columns}, {leaderboardExpanded ? "1fr" : "315px"})"
          class="
            grid
            gap-6 justify-start"
        >
          {#each currentLeaderboard.leaderboards as { displayName, values, nullCount }, i (displayName)}
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
