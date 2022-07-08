<script lang="ts">
  import BarAndLabel from "$lib/components/viz/BarAndLabel.svelte";
  import { formatInteger } from "$lib/util/formatters";
  import { cubicIn } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import type { Readable } from "svelte/types/runtime/store";
  import { formatBigNumberPercentage } from "$lib/util/formatters";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import { getMeasureFieldNameByIdAndIndex } from "$lib/redux-store/measure-definition/measure-definition-readables";

  export let metricsDefId: string;
  export let measureId: string;
  export let index: number;

  let bigNumberEntity: Readable<BigNumberEntity>;
  $: bigNumberEntity = getBigNumberById(metricsDefId);
  let measureField: Readable<string>;
  $: measureField = getMeasureFieldNameByIdAndIndex(measureId, index);

  const metricFormatters = {
    simpleSummable: formatInteger,
  };
  let bigNumber;
  $: bigNumber = $bigNumberEntity?.bigNumbers?.[$measureField] ?? 0;
  const bigNumberTween = tweened(0, {
    duration: 1000,
    delay: 200,
    easing: cubicIn,
  });
  $: bigNumberTween.set(bigNumber);
  let referenceValue: number;
  $: referenceValue = $bigNumberEntity?.referenceValues?.[$measureField] ?? 0;
</script>

<div class="w-full rounded text-lg">
  {$bigNumberTween}
</div>
