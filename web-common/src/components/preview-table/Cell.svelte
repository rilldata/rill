<script lang="ts">
  import FormattedDataType from "../data-types/FormattedDataType.svelte";

  export let value: string | boolean | undefined;
  export let type: string;
</script>

<div
  class:!text-right={type === "CODE_INT64" || type === "CODE_FLOAT64"}
  class="text-left px-6 truncate w-full"
>
  <FormattedDataType {type} {value} />
</div>
