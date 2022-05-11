<script>

    import Close from "$lib/components/icons/Close.svelte";
import Leaderboard from "./LeaderboardFeature.svelte";

    /** remove this before we componentize anything. */
    const files = import.meta.globEager('./data/*.json');
    const leaderboardSet = Object.keys(files).map(fileName => {
        const leaderboard = files[fileName].default;
            return [fileName, leaderboard];
    })

    /**
     * get the current leaderboard element.
     */
    let currentLeaderboard = leaderboardSet[0][1];
    let activeValues = {};

    function initializeActiveValues(leaderboards) {
        return leaderboards.reduce((acc, leaderboard) => {
        acc[leaderboard.displayName] = [];
        return acc;
    }, {});
    }

    function clearAllFilters() {
        activeValues = initializeActiveValues(currentLeaderboard.leaderboards); 
    }

    $: activeValues = initializeActiveValues(currentLeaderboard.leaderboards);

    $: anythingSelected = Object.keys(activeValues).some(key => {
        return activeValues[key]?.length;
    })
</script>

<div class="w-screen min-h-screen bg-white p-8">
    <section style:width="{315 * 3 + 2 * 32}px">
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
                <button 
                    on:click={clearAllFilters}
                >
                    clear all filters <Close />
                </button>
            {/if}
        </div>
    </header>
    <div class="grid grid-cols-3 gap-8 justify-start w-max">
        {#each currentLeaderboard.leaderboards as {displayName, values, nullCount }}
            <Leaderboard
                on:select-item={(event) => {
                    activeValues[displayName];
                    if (!(activeValues[displayName].includes(event.detail))) {
                        activeValues[displayName] = [...activeValues[displayName], event.detail]
                    } else {
                        activeValues[displayName] = activeValues[displayName].filter(b => b !== event.detail);
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
    {:else}
        <p>
            <b>No leaderboards present.</b>
        </p>
        <p style:width="600px">
            Run <code class="italic text-blue-600">node ./scripts/dev/generate-leaderboards.js path/to/stage.db</code>
            to generate example leaderboard data.
        </p>
    {/if}
    </section>
</div>
