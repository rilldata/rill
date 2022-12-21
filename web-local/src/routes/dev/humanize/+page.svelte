<script>
  import {
    applyScaleOnValues,
    determineScaleForValues,
  } from "@rilldata/web-local/lib/util/humanize-numbers";

  let maxValue = 1000000000;
  let numSamples = 250;

  function randomBoxMuller() {
    let u = 0,
      v = 0;
    while (u === 0) u = Math.random(); //Converting [0,1) to (0,1)
    while (v === 0) v = Math.random();
    let num = Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);
    num = num / 10.0 + 0.5; // Translate to 0 -> 1
    if (num > 1 || num < 0) return randomBoxMuller(); // resample between 0 and 1
    return num;
  }

  function getNormalColumn(totalCount) {
    let columnTwo = [];
    for (let i = 0; i < numSamples; i++) {
      columnTwo.push(randomBoxMuller());
    }
    const columnTwoSum = columnTwo.reduce((a, b) => a + b, 0);
    const multiplier = totalCount / columnTwoSum;

    columnTwo = columnTwo.map((v) => Math.floor(v * multiplier));
    columnTwo = columnTwo.sort((a, b) => b - a);
    return columnTwo;
  }

  let columnOne;
  $: {
    let value = maxValue ? maxValue : 1000000000;
    columnOne = [];
    for (let i = 0; i < numSamples; i++) {
      columnOne.push(value < 1 ? 1 : value);
      value = Math.floor(value / 2);
    }
  }

  $: totalCount = columnOne.reduce((a, b) => a + b, 0);

  $: columnTwo = getNormalColumn(totalCount);

  $: commonScale = determineScaleForValues(columnOne.concat(columnTwo));
  $: colOneFormatted = applyScaleOnValues(columnOne, commonScale);
  $: colTwoFormatted = applyScaleOnValues(columnTwo, commonScale);
</script>

<div class="flex">
  <div class="w-40 mr-14">
    <div class="p-3">
      <span class="m-1">Max value</span>
      <input
        type="number"
        class="p-1 border-2"
        label="Max value"
        bind:value={maxValue}
      />
    </div>
  </div>

  <div class="w-40">
    <div class="p-3">
      <span class="m-1">Number of Samples</span>
      <input
        type="number"
        class="p-1 border-2"
        label="Max value"
        bind:value={numSamples}
      />
    </div>
  </div>
</div>

<div class="grid grid-cols-2 w-max gap-x-10">
  <div class="text-lg text-left pt-5 px-2 mx-5">
    <div class="text-slate-700 w-60 m-auto text-center text-sm p-3">
      Start with max value and half until value becomes 1
    </div>
    {#each columnOne as val, i}
      <div>
        <span class="inline-block w-20 px-2">{colOneFormatted[i]}</span>
        <span class="w-40 float-right text-right">{val}</span>
      </div>
    {/each}
  </div>
  <div class="text-lg text-left pt-5 px-2 mx-5">
    <div class="text-slate-700 w-60 m-auto text-center text-sm p-3">
      Normal distribution with sum equal to that of columnn one
    </div>
    {#each columnTwo as val, i}
      <div>
        <span class="w-20 px-2">{colTwoFormatted[i]}</span>
        <span class="w-40 float-right text-right">{val}</span>
      </div>
    {/each}
  </div>
</div>
