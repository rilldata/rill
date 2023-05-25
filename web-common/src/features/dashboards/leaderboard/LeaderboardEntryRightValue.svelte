<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { humanizeDataType } from "../humanize-numbers";
  import { FormattedDataType } from "../../../components/data-types";

  export let label;

  export let value;
  export let comparisonValue;
  export let showComparisonForThisValue = false;
  export let formatPreset;
  export let formattedValue: string;

  $: if (label == "Facebook") {
    console.log(
      "LeaderboardEntryRightValue - saw Facebook",
      "value",
      value,
      "comparisonValue",
      comparisonValue,
      "formattedValue",
      formattedValue
      // "rowCount",
      // rowCount
    );
  }
</script>

<div class="flex items-baseline gap-x-1">
  {#if showComparisonForThisValue && comparisonValue !== undefined && comparisonValue !== null}
    <span
      class="inline-block opacity-50"
      transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
    >
      {humanizeDataType(comparisonValue, formatPreset)}
      →
    </span>
  {/if}
  <FormattedDataType type="INTEGER" value={formattedValue || value} />
</div>
