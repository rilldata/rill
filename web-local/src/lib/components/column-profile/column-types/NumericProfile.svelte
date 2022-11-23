<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import NullPercentageSpark from "../data-graphics/sparks/NullPercentageSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";

  import TopK from "../data-graphics/details/TopK.svelte";

  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
  import NumericPlot from "../data-graphics/details/NumericPlot.svelte";
  import NumericSpark from "../data-graphics/sparks/NumericSpark.svelte";
  export let columnName: string;
  export let objectName: string;
  export let type: string;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let columns: string;
  $: summarySize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];
  $: if (hideNullPercentage) {
    columns = `${summarySize}px`;
  } else {
    columns = `${summarySize}px ${COLUMN_PROFILE_CONFIG.nullPercentageWidth}px`;
  }

  let active = false;
</script>

<ProfileContainer
  on:select={() => {
    active = !active;
  }}
  on:shift-click={() =>
    copyToClipboard(columnName, `copied ${columnName} to clipboard`)}
  {active}
  emphasize={active}
  {hideRight}
>
  <DataTypeIcon {type} slot="icon" />
  <div slot="left">{columnName}</div>
  <div slot="right" class="grid" style:grid-template-columns={columns}>
    <div>
      <NumericSpark {compact} {objectName} {columnName} />
    </div>
    {#if !hideNullPercentage}
      <NullPercentageSpark {objectName} {columnName} />
    {/if}
  </div>

  <div slot="details">
    <NumericPlot {objectName} {columnName} />
    <TopK {objectName} {columnName} />
  </div>
</ProfileContainer>
