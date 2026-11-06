<script lang="ts">
  import DashboardMetricsDraggableList from "../DashboardMetricsDraggableList.svelte";
  import { buildTagIndex } from "../tag-utils";
  import * as Tooltip from "../../tooltip-v2";
  import CellInspector from "../../CellInspector.svelte";
  import DashboardShortcuts from "@rilldata/web-common/features/dashboards/shortcuts/DashboardShortcuts.svelte";

  export let type: "measure" | "dimension" = "measure";
  export let withTags = true;
  export let selectedItems = ["requests", "spend", "bids"];
  export let onChange: (items: string[]) => void = () => {};
  const allItems = [
    { name: "requests", displayName: "Requests", tags: ["Traffic"] },
    { name: "spend", displayName: "Spend", tags: ["Revenue"] },
    { name: "bids", displayName: "Bids", tags: ["Traffic"] },
  ];
  $: tagIndex = buildTagIndex(
    withTags
      ? allItems
      : allItems.map(({ name, displayName }) => ({ name, displayName })),
  );
</script>

<Tooltip.Provider>
  <DashboardMetricsDraggableList
    {type}
    {allItems}
    {tagIndex}
    {selectedItems}
    onSelectedChange={(items) => {
      selectedItems = items;
      onChange(items);
    }}
  />
  <DashboardShortcuts />
  <CellInspector />
</Tooltip.Provider>
