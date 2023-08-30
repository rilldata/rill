<script lang="ts">
  import { getColumnsProfileStore } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import type { ColumnProfileData } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import TimestampDetail from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampDetail.svelte";
  import TimestampSpark from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import WithParentClientRect from "@rilldata/web-common/components/data-graphic/functional-components/WithParentClientRect.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { TIMESTAMP_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
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

  const columnsProfile = getColumnsProfileStore();

  let columnProfileData: ColumnProfileData;
  $: columnProfileData = $columnsProfile.profiles[columnName];

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
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div class={TIMESTAMP_TOKENS.textClass} slot="summary">
    <WithParentClientRect let:rect>
      <TimestampSpark
        bottom={4}
        color={"currentColor"}
        data={columnProfileData?.timeSeriesSpark ?? []}
        height={18}
        top={4}
        width={rect?.width || 400}
        xAccessor="ts"
        yAccessor="count"
      />
    </WithParentClientRect>
  </div>
  <NullPercentageSpark
    nullCount={columnProfileData?.nullCount}
    slot="nullity"
    totalRows={$columnsProfile?.tableRows}
    {type}
  />

  <div slot="details">
    <div class="pl-8 py-4" style:height="{timestampDetailHeight + 64 + 28}px">
      <WithParentClientRect let:rect>
        {#if columnProfileData?.timeSeriesData?.length}
          <TimestampDetail
            width={rect?.width - 56 || 400}
            mouseover={true}
            height={timestampDetailHeight}
            data={columnProfileData?.timeSeriesData ?? []}
            spark={columnProfileData?.timeSeriesSpark ?? []}
            rollupTimeGrain={columnProfileData?.estimatedRollupInterval}
            estimatedSmallestTimeGrain={columnProfileData?.smallestTimeGrain}
            xAccessor="ts"
            yAccessor="count"
          />
        {/if}
      </WithParentClientRect>
    </div>
  </div>
</ProfileContainer>
