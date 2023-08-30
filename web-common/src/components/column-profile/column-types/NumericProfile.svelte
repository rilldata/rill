<script lang="ts">
  import { getColumnsProfileStore } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import type { ColumnProfileData } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { INTERVALS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import NumericPlot from "./details/NumericPlot.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";
  import NumericSpark from "./sparks/NumericSpark.svelte";

  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let active = false;

  const columnsProfile = getColumnsProfileStore();

  let columnProfileData: ColumnProfileData;
  $: columnProfileData = $columnsProfile.profiles[columnName];

  $: histogramData = columnProfileData?.histogram;

  $: summary = columnProfileData?.descriptiveStatistics;

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  /** if we have a singleton where all summary information is the same, let's construct a single bin. */
  // TODO: move this to their own methods
  $: if (
    summary?.min !== undefined &&
    summary?.min === summary?.max &&
    $columnsProfile?.tableRows !== undefined
  ) {
    const boundaries = 10;
    histogramData = [
      // add 4 more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: -boundaries + i,
          count: 0,
          high: summary?.min - (boundaries - i - 1),
          low: summary?.min - (boundaries - i),
        };
      }),
      {
        bucket: boundaries,
        count: $columnsProfile?.tableRows,
        low: summary?.min,
        high: summary?.min + 1,
      },
      // add more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: boundaries + i + 1,
          count: 0,
          low: summary?.min + i,
          high: summary?.min + i + 1,
        };
      }),
    ];
  }
</script>

<ProfileContainer
  {active}
  {compact}
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
  <NumericSpark {compact} data={histogramData} slot="summary" {type} />
  <NullPercentageSpark
    nullCount={columnProfileData?.nullCount}
    slot="nullity"
    totalRows={$columnsProfile?.tableRows}
    {type}
  />
  <div
    class="pl-10 pr-4 py-4"
    class:hidden={INTERVALS.has(type)}
    slot="details"
  >
    <NumericPlot
      data={histogramData}
      rug={columnProfileData?.rugHistogram}
      {summary}
      topK={columnProfileData?.topK}
      totalRows={$columnsProfile?.tableRows}
      {type}
    />
  </div>
</ProfileContainer>
