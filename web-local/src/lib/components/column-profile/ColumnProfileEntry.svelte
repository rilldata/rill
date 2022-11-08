<script lang="ts">
  import { DataTypeIcon } from "../data-types";
  import ColumnEntry from "./ColumnEntry.svelte";

  import notificationStore from "./ColumnProfile.svelte";
  import ColumnProfileDetails from "./ColumnProfileDetails.svelte";
  import ColumnProfileTitle from "./ColumnProfileTitle.svelte";
  import ColumnSummaryMiniPlots from "./ColumnSummaryMiniPlots.svelte";

  export let name;
  export let type;
  export let summary;
  export let totalRows;
  export let nullCount;
  export let example;
  export let entityId;
  export let view: "summaries" | "example" = "summaries"; // summaries, example
  export let containerWidth: number;

  export let indentLevel = 1;

  export let hideRight = false;
  export let hideNullPercentage = false;
  export let compactBreakpoint = 350;

  let active = false;
</script>

<!-- pl-10 -->
<ColumnEntry
  left={indentLevel === 1 ? 10 : 4}
  {hideRight}
  {active}
  emphasize={active}
  on:shift-click={async () => {
    await navigator.clipboard.writeText(name);
    notificationStore.send({
      message: `copied column name "${name}" to clipboard`,
    });
  }}
  on:select={async () => {
    // we should only allow activation when there are rows present.
    if (totalRows) {
      active = !active;
    }
  }}
>
  <DataTypeIcon slot="icon" {type} />

  <ColumnProfileTitle slot="left" {...{ name, type, totalRows, active }} />

  <ColumnSummaryMiniPlots
    slot="right"
    {...{
      type,
      summary,
      totalRows,
      nullCount,
      example,
      view,
      containerWidth,
      hideNullPercentage,
      compactBreakpoint,
    }}
  />

  <svelte:fragment slot="context-button">
    <slot name="context-button" />
  </svelte:fragment>

  <ColumnProfileDetails
    slot="details"
    {...{
      active,
      type,
      summary,
      totalRows,
      containerWidth,
      indentLevel,
      name,
      entityId,
    }}
  />
</ColumnEntry>
