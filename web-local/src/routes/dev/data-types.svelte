<script>
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";

  const examples = [
    { value: 10.534234, type: "FLOAT", label: "float", comment: "" },
    { value: 53243, type: "INTEGER", label: "integer" },
    { value: "test text", type: "VARCHAR", label: "varchar" },
    { value: new Date(), type: "TIMESTAMP", label: "timestamp" },
    { value: "1985-12-04", type: "DATE", label: "date" },
    {
      value: { months: 0, days: 0, micros: 599300 },
      type: "INTERVAL",
      label: "interval (ms)",
    },
    {
      value: { months: 0, days: 0, micros: 10503020202 },
      type: "INTERVAL",
      label: "interval (inner-day)",
    },
    {
      value: { months: 0, days: 1, micros: 12345679039 },
      type: "INTERVAL",
      label: "interval (about a day)",
    },
    {
      value: { months: 0, days: 3, micros: 10503020202 },
      type: "INTERVAL",
      label: "interval (a few days)",
    },
    {
      value: { months: 0, days: 4 * 7 + 1, micros: 0 },
      type: "INTERVAL",
      label: "interval (a few weeks)",
    },
    {
      value: { months: 0, days: 364, micros: 34939439 },
      type: "INTERVAL",
      label: "interval (almost a year)",
    },
    {
      value: { months: 0, days: 364 * 2.43, micros: 34939439 },
      type: "INTERVAL",
      label: "interval (a few years)",
    },
    {
      value: { months: 0, days: 364 * 54.43, micros: 34939439 },
      type: "INTERVAL",
      label: "interval (many years)",
    },
    { value: null, type: "", label: "nulls" },
  ];
</script>

<h1 class="pb-4">Data Type Value Representations</h1>

<div class="grid grid-cols-3 w-max gap-x-4">
  {#each examples as { value, type, label }}
    <div class="flex items-center gap-x-2 pl-2 pr-2">
      <DataTypeIcon {type} />
      {label}
    </div>
    <div class="text-right pl-2 pr-2">
      <FormattedDataType {value} {type} />
    </div>
    <div class="text-right bg-gray-800 pl-2 pr-2">
      <FormattedDataType {value} {type} dark />
    </div>
  {/each}
</div>
