<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";

  import ColumnCardinalitySpark from "../data-graphics/sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "../data-graphics/sparks/NullPercentageSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import TopK from "../data-graphics/details/TopK.svelte";

  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
  export let columnName: string;
  export let objectName: string;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let mode: "example" | "summary" = "summary";

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
  <DataTypeIcon type="VARCHAR" slot="icon" />
  <div slot="left">{columnName}</div>
  <div
    slot="right"
    class="grid"
    style:hidden={hideRight}
    style:grid-template-columns={mode === "summary" ? columns : "auto"}
  >
    {#if mode === "summary"}
      <div>
        <ColumnCardinalitySpark {compact} {objectName} {columnName} />
      </div>
      {#if !hideNullPercentage}
        <NullPercentageSpark {objectName} {columnName} />
      {/if}
    {:else}
      example
    {/if}
  </div>

  <div slot="details">
    <TopK {objectName} {columnName} />
  </div>
</ProfileContainer>
