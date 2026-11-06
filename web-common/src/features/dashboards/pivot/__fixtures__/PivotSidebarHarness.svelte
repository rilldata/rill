<script lang="ts">
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { buildTagIndex } from "@rilldata/web-common/components/menu/tag-utils";
  import { DEFAULT_STORE_KEY } from "../../state-managers/state-managers";
  import { metricsExplorerStore } from "../../stores/dashboard-stores";
  import { initAdBidsInStore } from "../../stores/test-data/helpers";
  import { AD_BIDS_EXPLORE_NAME } from "../../stores/test-data/data";
  import PivotSidebar from "../PivotSidebar.svelte";
  import { PivotChipType } from "../types";

  export let flat = false;
  const name = AD_BIDS_EXPLORE_NAME;
  initAdBidsInStore();
  metricsExplorerStore.setPivotRows(name, []);
  metricsExplorerStore.setPivotColumns(name, []);
  setContext(DEFAULT_STORE_KEY, { exploreName: writable(name) });
  const measures = [
    { id: "requests", title: "Requests", type: PivotChipType.Measure },
    { id: "spend", title: "Spend", type: PivotChipType.Measure },
  ];
  const dimensions = [
    { id: "country", title: "Country", type: PivotChipType.Dimension },
    { id: "region", title: "Region", type: PivotChipType.Dimension },
  ];
  const measureTagIndex = buildTagIndex(
    measures.map((item) => ({ name: item.id, tags: ["Traffic"] })),
  );
  const dimensionTagIndex = buildTagIndex(
    dimensions.map((item) => ({ name: item.id, tags: ["Location"] })),
  );
  const combinedTagIndex = buildTagIndex(
    [...measures, ...dimensions].map((item) => ({
      name: item.id,
      tags: [item.type === PivotChipType.Measure ? "Traffic" : "Location"],
    })),
  );
  $: state = $metricsExplorerStore.entities[name].pivot;
  $: used = new Set([...state.rows, ...state.columns].map((item) => item.id));
</script>

<Tooltip.Provider>
  <PivotSidebar
    pivotState={{ ...state, tableMode: flat ? "flat" : "nest" }}
    measures={measures.filter((item) => !used.has(item.id))}
    dimensions={dimensions.filter((item) => !used.has(item.id))}
    {combinedTagIndex}
    {dimensionTagIndex}
    {measureTagIndex}
    setRows={(rows) => metricsExplorerStore.setPivotRows(name, rows)}
    setColumns={(columns) =>
      metricsExplorerStore.setPivotColumns(name, columns)}
    timeControlsForPillActions={{
      timeStart: "2026-10-01",
      timeEnd: "2026-11-01",
    }}
  />
</Tooltip.Provider>
