<script>
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "$lib/util/humanize-numbers";

  const examples = [
    { value: 10.534234, type: NicelyFormattedTypes.DECIMAL },
    { value: 3242340.534234, type: NicelyFormattedTypes.DECIMAL },
    { value: 53243, type: NicelyFormattedTypes.COUNT },
    { value: 1253243, type: NicelyFormattedTypes.COUNT },
    { value: 23153243, type: NicelyFormattedTypes.COUNT },
    { value: 1121153243, type: NicelyFormattedTypes.COUNT },
    { value: 132.224, type: NicelyFormattedTypes.CURRENCY },
    { value: 2132.3114, type: NicelyFormattedTypes.CURRENCY_WITH_NAME },
    { value: -423241.3451, type: NicelyFormattedTypes.ACCOUNTING_CURRENCY },
    { value: 0.94995328, type: NicelyFormattedTypes.PERCENTAGE },
  ];
</script>

<h1 class="pb-4">Data Formatting Representations</h1>

<!-- <div class="text-2xl">Individual value formatting</div> -->
<div class="grid grid-cols-2 w-max gap-x-5">
  {#each examples as { value, type, label }}
    {@const formatted = humanizeDataType(value, type)}
    <div class="text-lg text-left pt-5 pl-2 pr-2">
      <div>
        Label: {type} <br />
        Original: <b>{value}</b> <br />
        Formatted:<b> {formatted}</b>
      </div>
    </div>
  {/each}
</div>
