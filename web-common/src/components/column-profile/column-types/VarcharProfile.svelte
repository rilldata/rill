<script lang="ts">
  import { getColumnsProfileStore } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import type { ColumnProfileData } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import TopK from "./details/TopK.svelte";
  import ColumnCardinalitySpark from "./sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let columnName: string;
  export let objectName: string;
  export let example: any;
  export let type: string;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let mode: "example" | "summaries" = "summaries";

  let active = false;

  let topKLimit = 15;

  const columnsProfile = getColumnsProfileStore();
  let columnProfileData: ColumnProfileData;
  $: columnProfileData = $columnsProfile.profiles[columnName];
  $: tableRows = $columnsProfile.tableRows;

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }
</script>

<ProfileContainer
  {active}
  emphasize={active}
  {example}
  {hideNullPercentage}
  {hideRight}
  isFetching={columnProfileData?.isFetching}
  {mode}
  on:select={toggleColumnProfile}
  on:shift-click={() =>
    copyToClipboard(columnName, `copied ${columnName} to clipboard`)}
  {type}
>
  <ColumnProfileIcon
    isFetching={columnProfileData?.isFetching}
    slot="icon"
    {type}
  />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    cardinality={columnProfileData?.cardinality}
    {compact}
    slot="summary"
    totalRows={tableRows}
  />
  <NullPercentageSpark
    nullCount={columnProfileData?.nullCount}
    slot="nullity"
    totalRows={tableRows}
    {type}
  />

  <div
    class="pl-10 pr-4 py-4"
    slot="details"
    style:min-height="{Math.min(topKLimit, columnProfileData?.cardinality) *
      18 +
      42 || 250}px"
  >
    <div>
      <TopK
        k={topKLimit}
        topK={columnProfileData?.topK}
        totalRows={$columnsProfile?.tableRows}
        {type}
      />
    </div>
  </div>
</ProfileContainer>
