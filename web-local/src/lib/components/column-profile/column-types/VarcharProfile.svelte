<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";

  import TopK from "../data-graphics/details/TopK.svelte";
  import ColumnCardinalitySpark from "../data-graphics/sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "../data-graphics/sparks/NullPercentageSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";

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
  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark slot="summary" {compact} {objectName} {columnName} />
  <NullPercentageSpark slot="nullity" {objectName} {columnName} />

  <div slot="details" class="px-4">
    <TopK {objectName} {columnName} />
  </div>
</ProfileContainer>
