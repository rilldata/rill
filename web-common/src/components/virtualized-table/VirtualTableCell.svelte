<script lang="ts">
  import FormattedDataType from "../../components/data-types/FormattedDataType.svelte";

  export let value: string | boolean | number | null | undefined;
  export let type: string | undefined;
  export let selected: boolean;
  export let sorted: boolean;
  export let formattedValue: string | undefined;

  $: finalValue = formattedValue ?? value;
</script>

<div
  class:sorted
  class:selected
  class:!text-left={type === "VARCHAR"}
  class="text-right px-6 truncate w-full"
>
  <FormattedDataType {type} value={finalValue} isNull={value === null} />
</div>
