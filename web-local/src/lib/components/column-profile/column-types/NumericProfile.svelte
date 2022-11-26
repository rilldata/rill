<script lang="ts">
  import NullPercentageSpark from "../data-graphics/sparks/NullPercentageSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";

  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
  import NumericPlot from "../data-graphics/details/NumericPlot.svelte";
  import NumericSpark from "../data-graphics/sparks/NumericSpark.svelte";
  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

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
  {hideNullPercentage}
  {compact}
  {mode}
  {example}
  {type}
>
  <DataTypeIcon {type} slot="icon" />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>
  <NumericSpark slot="summary" {compact} {objectName} {columnName} />
  <NullPercentageSpark slot="nullity" {objectName} {columnName} />
  <div slot="details" class="px-4">
    <NumericPlot {objectName} {columnName} />
  </div>
</ProfileContainer>
