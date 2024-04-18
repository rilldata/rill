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
  class:!justify-start={type === "VARCHAR" || type === "CODE_STRING"}
  class=" px-6 size-full flex items-center"
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
