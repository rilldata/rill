<script lang="ts">
  import TimestampDetail from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampDetail.svelte";
  import TimestampSpark from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import WithParentClientRect from "@rilldata/web-common/components/data-graphic/functional-components/WithParentClientRect.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { TIMESTAMP_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getTimeSeriesAndSpark,
    isFetching,
  } from "../queries";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let timestampDetailHeight = 160;

  let active = false;

  /** queries used to power the different plots */
  $: nullPercentage = getNullPercentage(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );

  $: timeSeries = getTimeSeriesAndSpark(
    $runtimeStore?.instanceId,
    objectName,
    columnName,
    active
  );

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($timeSeries, $nullPercentage);
</script>

<ProfileContainer
  isFetching={fetchingSummaries}
  {active}
  {compact}
  emphasize={active}
  {example}
  {hideNullPercentage}
  {hideRight}
  {mode}
  on:select={toggleColumnProfile}
  on:shift-click={() =>
    copyToClipboard(columnName, `copied ${columnName} to clipboard`)}
  {type}
>
  <ColumnProfileIcon slot="icon" {type} isFetching={fetchingSummaries} />
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div class={TIMESTAMP_TOKENS.textClass} slot="summary">
    <WithParentClientRect let:rect>
      <TimestampSpark
        bottom={4}
        color={"currentColor"}
        data={$timeSeries?.spark}
        height={18}
        top={4}
        width={rect?.width || 400}
        xAccessor="ts"
        yAccessor="count"
      />
    </WithParentClientRect>
  </div>
  <NullPercentageSpark
    isFetching={fetchingSummaries}
    nullCount={$nullPercentage?.nullCount}
    slot="nullity"
    totalRows={$nullPercentage?.totalRows}
    {type}
  />

  <div slot="details">
    <div class="pl-8 py-4" style:height="{timestampDetailHeight + 64 + 28}px">
      <WithParentClientRect let:rect>
        {#if $timeSeries?.data?.length}
          <TimestampDetail
            width={rect?.width - 56 || 400}
            mouseover={true}
            height={timestampDetailHeight}
            {type}
            data={$timeSeries?.data}
            spark={$timeSeries?.spark}
            rollupTimeGrain={$timeSeries?.estimatedRollupInterval?.interval}
            estimatedSmallestTimeGrain={$timeSeries?.smallestTimegrain}
            xAccessor="ts"
            yAccessor="count"
          />
        {/if}
      </WithParentClientRect>
    </div>
  </div>
</ProfileContainer>
