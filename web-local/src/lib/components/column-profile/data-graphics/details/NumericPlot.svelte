<script lang="ts">
  import { WithParentClientRect } from "../../../data-graphic/functional-components";
  import NumericHistogram from "../../../viz/histogram/NumericHistogram.svelte";
  import OutlierHistogram from "../../../viz/histogram/OutlierHistogram.svelte";

  export let data;
  export let rug;
  export let summary;
  export let plotPad = 24;
</script>

<WithParentClientRect let:rect>
  {#if data && summary}
    <NumericHistogram
      width={(rect?.width || 400) - plotPad}
      height={65}
      {data}
      min={summary?.min}
      qlow={summary?.q25}
      median={summary?.q50}
      qhigh={summary?.q75}
      mean={summary?.mean}
      max={summary?.max}
    />
  {/if}
  {#if rug && rug?.length}
    <OutlierHistogram
      width={(rect?.width || 400) - plotPad}
      height={15}
      data={rug}
      mean={summary?.mean}
      sd={summary?.sd}
      min={summary?.min}
      max={summary?.max}
    />
  {/if}
</WithParentClientRect>
