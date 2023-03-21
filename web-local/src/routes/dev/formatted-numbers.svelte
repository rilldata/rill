<script>
  import {
    humanizeDataType_legacy,
    humanizeGroupValues,
    NicelyFormattedTypes,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";

  const examples = [
    { value: 3242340.534234, type: NicelyFormattedTypes.HUMANIZE },
    { value: 53243, type: NicelyFormattedTypes.HUMANIZE },
    { value: 1253243, type: NicelyFormattedTypes.HUMANIZE },
    { value: 23153243, type: NicelyFormattedTypes.HUMANIZE },
    { value: 1121153243, type: NicelyFormattedTypes.HUMANIZE },
    { value: 132.224, type: NicelyFormattedTypes.CURRENCY },
    { value: 0.94995328, type: NicelyFormattedTypes.PERCENTAGE },
  ];

  const table = [
    { value: 614474, label: "Noise - Residential" },
    { value: 261361, label: "Blocked Driveway" },
    { value: 224188, label: "Illegal Parking" },
    { value: 195016, label: "Water System" },
    { value: 172282, label: "Damaged Tree" },
    { value: 161099, label: "Street Condition" },
    { value: 142032, label: "Noise - Street/Sidewalk" },
    { value: 119344, label: "Sewer" },
    { value: 115530, label: "Noise" },
    { value: 103443, label: "Rodent" },
    { value: 101252, label: "Noise - Commercial" },
    { value: 85557, label: "Sidewalk Condition" },
    { value: 82029, label: "Broken Muni Meter" },
    { value: 80491, label: "Noise - Vehicle" },
    { value: 75880, label: "Taxi Complaint" },
  ];

  const secondTable = [
    { value: 31212614474, label: "Noise - Residential" },
    { value: 1232261361, label: "Blocked Driveway" },
    { value: 242332333, label: "Illegal Parking" },
    { value: 134334232, label: "Water System" },
    { value: 124232442, label: "Damaged Tree" },
    { value: 45343534, label: "Street Condition" },
    { value: 34343344, label: "Noise - Street/Sidewalk" },
    { value: 23454545, label: "Sewer" },
    { value: 8534344, label: "Noise" },
    { value: 8353432, label: "Rodent" },
    { value: 634534, label: "Noise - Commercial" },
    { value: 95343, label: "Sidewalk Condition" },
    { value: 82029, label: "Broken Muni Meter" },
    { value: 80491, label: "Noise - Vehicle" },
    { value: 75880, label: "Taxi Complaint" },
  ];

  $: humanizedTable = humanizeGroupValues(table, NicelyFormattedTypes.HUMANIZE);
  $: humanizedTable2 = humanizeGroupValues(
    secondTable,
    NicelyFormattedTypes.HUMANIZE
  );

  $: humanizedTableCurrency = humanizeGroupValues(
    table,
    NicelyFormattedTypes.CURRENCY
  );
</script>

<h1 class="pb-4">Data Formatting Representations</h1>

<!-- <div class="text-2xl">Individual value formatting</div> -->
<div class="grid grid-cols-4 w-max gap-x-5">
  {#each examples as { value, type, label }}
    {@const formatted = humanizeDataType_legacy(value, type)}
    <div class="text-lg text-left pt-5 pl-2 pr-2">
      <div>
        Label: {type} <br />
        Original: <b>{value}</b> <br />
        Formatted:<b> {formatted}</b>
      </div>
    </div>
  {/each}
  <div class="text-lg text-left pt-5 pl-2 pr-2">
    {#each humanizedTable as { __formatted_value, label }}
      <div>
        <span>{label}</span>
        <span class="font-bold float-right">{__formatted_value}</span>
      </div>
    {/each}
  </div>
  <div class="text-lg text-left pt-5 pl-2 pr-2">
    {#each humanizedTable2 as { __formatted_value, label }}
      <div>
        <span>{label}</span>
        <span class="font-bold float-right">{__formatted_value}</span>
      </div>
    {/each}
  </div>
  <div class="text-lg text-left pt-5 pl-2 pr-2">
    {#each humanizedTableCurrency as { __formatted_value, label }}
      <div>
        <span>{label}</span>
        <span class="font-bold float-right">{__formatted_value}</span>
      </div>
    {/each}
  </div>
</div>
