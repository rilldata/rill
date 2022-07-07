<script lang="ts">
  import BarAndLabel from "$lib/components/BarAndLabel.svelte";
  import { formatInteger } from "$lib/util/formatters";
  import { cubicIn } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import type { Readable } from "svelte/types/runtime/store";
  import { formatBigNumberPercentage } from "$lib/util/formatters.js";
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
  $: console.log(bigNumber, referenceValue);
</script>

<div class="w-full rounded">
  <BarAndLabel
    justify="stretch"
    color="bg-blue-200"
    value={referenceValue === 0 ? 0 : bigNumber / referenceValue}
  >
    <div
      style:grid-template-columns="auto auto"
      class="grid items-center gap-x-2 w-full text-left pb-2 pt-2"
    >
      <div>
        {metricFormatters.simpleSummable(~~$bigNumberTween)}
      </div>
      <div class="font-normal text-gray-600 italic text-right">
        {#if $bigNumberTween && referenceValue}
          {formatBigNumberPercentage($bigNumberTween / referenceValue)}
        {/if}
      </div>
    </div>
  </BarAndLabel>
</div>
