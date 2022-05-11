<script>

    import Leaderboard from "./LeaderboardFeature.svelte";

    /** remove this before we componentize anything. */
    const files = import.meta.globEager('./data/*.json');
    const leaderboardSet = Object.keys(files).map(fileName => {
        const leaderboard = files[fileName].default;
            return [fileName, leaderboard];
    })

    /** these parameters control the total leaderboard length */
    const slice = 3;
    const seeMoreSlice = 15;

    /** bootstrapping some data and interactions */
    const leaderboards = [
        {
            displayName: 'Publishers',
            total: 80000,
            values: [
                {
                    label: 'Goofball TV',
                    value: 42400
                },
                {
                    label: 'Nexus Television',
                    value: 20100
                },
                {
                    label: 'YouFace',
                    value: 12000
                },
                {
                    label: 'etc. online',
                    value: 9200
                },
                {
                    label: 'something else',
                    value: 5300
                },
                {
                    label: 'connected tv',
                    value: 3000
                },
                {
                    label: 'Balogne Television',
                    value: 1000
                },
                {
                    label: 'HGWTFTV',
                    value: 400
                },
            ]
        }
    ]

    let activeLeaderboards = leaderboardSet.reduce((acc,v) => {
        acc[v.displayName] = []
        return acc;
    }, {});
    let seeMore = true;


    /**
     * get the current leaderboard element.
     */
    let currentLeaderboard = leaderboardSet[0][1];
    let activeValues = {};
    $: activeValues = currentLeaderboard.leaderboards.reduce((acc, leaderboard) => {
        acc[leaderboard.displayName] = [];
        return acc;
    }, {});
</script>

<div class="w-screen min-h-screen bg-gray-50 p-8">

<h1>{currentLeaderboard.displayName}</h1>
<select bind:value={currentLeaderboard}>
    {#each leaderboardSet as [file, leaderboard]}
        <option value={leaderboard}>{leaderboard.displayName}</option>
    {/each}
</select>

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
</div>