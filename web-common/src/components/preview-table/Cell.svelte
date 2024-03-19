<script lang="ts">
  import FormattedDataType from "../data-types/FormattedDataType.svelte";

  export let value: string | boolean | undefined;
  export let type: string;
  export let selected: boolean;
  export let sorted: boolean;
  export let formattedValue: string | null = null;
  export let nullable = false;
</script>

<div
  class:sorted
  class:selected
  class:!text-right={type === "CODE_INT64" || type === "CODE_FLOAT64"}
  class="text-left px-6 truncate w-full"
>
  <FormattedDataType {type} {value} />
</div>
