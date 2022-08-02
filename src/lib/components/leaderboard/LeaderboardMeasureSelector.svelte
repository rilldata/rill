<script lang="ts">
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import SelectMenu from "$lib/components/menu/SimpleSelectorMenu.svelte";
  import { setMeasureIdAndUpdateLeaderboard } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import type { Readable } from "svelte/store";
  import { crossfade, fly } from "svelte/transition";

  export let metricsDefId;

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
  }

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  function handleMeasureUpdate(measureID) {
    setMeasureIdAndUpdateLeaderboard(store.dispatch, metricsDefId, measureID);
  }

  function formatForSelector(measure) {
    return {
      ...measure,
      key: measure.id,
      main: measure.label?.length ? measure.label : measure.expression,
    };
  }

  /** this should be a single element */
  let selections = [];
  // reset selections based on the active leaderboard measure
  $: previous = { ...activeLeaderboardMeasure };
  let activeLeaderboardMeasure;

  $: activeLeaderboardMeasure =
    $measures?.length &&
    formatForSelector(
      $measures.find(
        (measure) => measure.id === $metricsExplorer?.leaderboardMeasureId
      )
    );

  const [send, receive] = crossfade({ fallback: fly, duration: 200 });

  /** this controls the animation direction */
  let dir = 1;

  $: options =
    $measures?.map((measure) => {
      let main = measure.label?.length ? measure.label : measure.expression;
      let description = main === measure.expression ? "" : measure.expression;
      return {
        ...measure,
        key: measure.id,
        main,
        description,
      };
    }) || [];
  $: selections = $measures ? [activeLeaderboardMeasure] : [];
</script>

<div class="flex flex-row items-center" style:grid-column-gap=".4rem">
  {#if $measures}
    <div>Dimension Leaders by</div>
    {#if $measures}
      <SelectMenu
        {options}
        {selections}
        alignment="end"
        on:select={(event) => {
          const key = event.detail[0].key;
          handleMeasureUpdate(key);
        }}
      />
    {/if}
    <!-- <SelectMenu
      alignment="end"
      options={}
      on:select={(event) => {
        const key = event.detail[0].key;
        /** set the direction based on the movement*/
        if (
          $measures.findIndex(
            (measure) => measure.id === activeLeaderboardMeasure.key
          ) > $measures.findIndex((measure) => measure.id === key)
        ) {
          dir = -1;
        } else {
          dir = 1;
        }
        handleMeasureUpdate(key);
      }}
      {selections}
      let:toggleMenu
      let:active
    >
      <button
        on:click={toggleMenu}
        class="font-bold grid grid-flow-col items-center gap-x-2 px-2 py-1 hover:bg-gray-200 {active
          ? 'bg-gray-200'
          : ''} rounded transition-color"
      >
        <div class="invisible ">
          {activeLeaderboardMeasure?.main}
        </div>
        {#key activeLeaderboardMeasure?.main}
          <div
            class="absolute "
            in:send|local={{
              key: activeLeaderboardMeasure.key,
              y: 8 * dir,
              duration: 200,
            }}
            out:receive|local={{
              key: activeLeaderboardMeasure.key,
              y: 8 * dir,
              duration: 200,
            }}
          >
            {activeLeaderboardMeasure?.main}
          </div>
        {/key}
        <div class=" -rotate-{active ? '180' : '0'} transition-transform">
          <CaretDownIcon />
        </div></button
      >
    </SelectMenu> -->
  {/if}
</div>
