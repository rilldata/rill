<script lang="ts">
  import TimestampDetail from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampDetail.svelte";
  import TimestampSpark from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { TIMESTAMP_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getTimeSeriesAndSpark,
    isFetching,
  } from "../queries";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let connector: string;
  export let database: string;
  export let databaseSchema: string;
  export let objectName: string;
  export let columnName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;
  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let enableProfiling = true;

  $: ({ instanceId } = $runtime);

  let timestampDetailHeight = 160;
  let active = false;
  let clientWidth: number;
  let secondWidth: number;

  /** queries used to power the different plots */
  $: nullPercentage = getNullPercentage(
    instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
  );

  $: timeSeries = getTimeSeriesAndSpark(
    instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
    active,
  );

  $: fetchingSummaries = isFetching($timeSeries, $nullPercentage);

  $: ({ data, spark } = $timeSeries);

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }
</script>

<ProfileContainer
  {active}
  {compact}
  emphasize={active}
  {example}
  {hideNullPercentage}
  {hideRight}
  isFetching={fetchingSummaries}
  {mode}
  onSelect={toggleColumnProfile}
  onShiftClick={() => copyToClipboard(columnName)}
  {type}
>
  <ColumnProfileIcon isFetching={fetchingSummaries} slot="icon" {type} />
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div class={TIMESTAMP_TOKENS.textClass} slot="summary" bind:clientWidth>
    <TimestampSpark
      bottom={4}
      color="currentColor"
      data={spark}
      height={18}
      top={4}
      width={clientWidth || 400}
    />
  </div>
  <NullPercentageSpark
    nullCount={$nullPercentage?.nullCount}
    slot="nullity"
    totalRows={$nullPercentage?.totalRows}
    {type}
  />

  <div slot="details">
    <div
      class="pl-8 py-4"
      style:height="{timestampDetailHeight + 64 + 28}px"
      bind:clientWidth={secondWidth}
    >
      {#if $timeSeries?.data?.length && $timeSeries?.estimatedRollupInterval?.interval && $timeSeries?.smallestTimegrain}
        <TimestampDetail
          width={secondWidth - 56 || 400}
          mouseover={true}
          height={timestampDetailHeight}
          {data}
          {spark}
          rollupTimeGrain={$timeSeries?.estimatedRollupInterval?.interval}
          estimatedSmallestTimeGrain={$timeSeries?.smallestTimegrain}
        />
      {/if}
    </div>
  </div>
</ProfileContainer>
