<script lang="ts">
  import NullPercentageSpark from "../data-graphics/sparks/NullPercentageSpark.svelte";
  import TimestampSpark from "../data-graphics/sparks/TimestampSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";

  import TopK from "../data-graphics/details/TopK.svelte";

  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let columns: string;

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
  {compact}
  {hideNullPercentage}
  {mode}
  {example}
  {type}
>
  <DataTypeIcon type="TIMESTAMP" slot="icon" />
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div slot="summary">
    <TimestampSpark
      height={18}
      top={4}
      bottom={4}
      xAccessor="ts"
      yAccessor="count"
      {objectName}
      {columnName}
    />
  </div>
  <NullPercentageSpark slot="nullity" {objectName} {columnName} />

  <div slot="details">
    <TopK {objectName} {columnName} />
  </div>
</ProfileContainer>
