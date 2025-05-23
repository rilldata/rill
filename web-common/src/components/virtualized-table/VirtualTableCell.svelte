<script lang="ts">
  import FormattedDataType from "../../components/data-types/FormattedDataType.svelte";

  export let value: unknown;
  export let type: string | undefined;
  export let selected: boolean;
  export let sorted: boolean;
  export let formattedValue: unknown;

  $: finalValue = (formattedValue ?? value) as
    | string
    | boolean
    | number
    | null
    | undefined;
</script>

<div
  class:sorted
  class:selected
  class:!justify-start={type === "VARCHAR" || type === "CODE_STRING"}
  class=" px-6 size-full flex items-center"
  data-cell-value={value?.toString() || ""}
>
  <p
    class="w-full truncate text-right"
    class:!text-left={type === "VARCHAR" || type === "CODE_STRING"}
  >
    <FormattedDataType
      truncate
      {type}
      value={finalValue}
      isNull={value === null}
    />
  </p>
</div>
