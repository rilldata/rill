<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import TimestampSpark from "../../data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  import { TIMESTAMP_TOKENS } from "@rilldata/web-local/lib/duckdb-data-types";
  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import TimestampDetail from "../../data-graphic/compositions/timestamp-profile/TimestampDetail.svelte";
  import WithParentClientRect from "../../data-graphic/functional-components/WithParentClientRect.svelte";
  import { DataTypeIcon } from "../../data-types";
  import Interval from "../../data-types/Interval.svelte";
  import { getNullPercentage, getTimeSeriesAndSpark } from "../queries";
  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

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
    columnName
  );
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
  <DataTypeIcon {type} slot="icon" />
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div slot="summary" class={TIMESTAMP_TOKENS.textClass}>
    <WithParentClientRect let:rect>
      <TimestampSpark
        area
        width={rect?.width || 400}
        height={18}
        top={4}
        bottom={4}
        xAccessor="ts"
        yAccessor="count"
        data={$timeSeries?.spark}
        color={"currentColor"}
      />
    </WithParentClientRect>
  </div>
  <NullPercentageSpark
    slot="nullity"
    nullCount={$nullPercentage?.nullCount}
    totalRows={$nullPercentage?.totalRows}
    {type}
  />

  <div slot="details">
    <div class="px-10 py-4">
      <WithParentClientRect let:rect>
        {#if $timeSeries?.data?.length}
          <TimestampDetail
            width={rect?.width - 56 || 400}
            mouseover={true}
            height={160}
            {type}
            data={$timeSeries?.data}
            spark={$timeSeries?.spark}
            interval={Interval[$timeSeries?.estimatedRollupInterval]}
            estimatedSmallestTimeGrain={$timeSeries?.smallestTimegrain}
            xAccessor="ts"
            yAccessor="count"
          />
        {/if}
      </WithParentClientRect>
    </div>
  </div>
</ProfileContainer>
