<script lang="ts">
  import FormattedDataType from "../../components/data-types/FormattedDataType.svelte";
  import { cellInspectorStore } from "../../features/dashboards/stores/cell-inspector-store";

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

  function handleMouseOver() {
    // Always update the value in the store, but don't change visibility
    cellInspectorStore.updateValue(
      value === null || value === undefined ? null : value.toString(),
    );
  }

  function handleFocus() {
    // Always update the value in the store, but don't change visibility
    cellInspectorStore.updateValue(
      value === null || value === undefined ? null : value.toString(),
    );
  }
</script>

<div
  role="cell"
  class:sorted
  class:selected
  class:!justify-start={type === "VARCHAR" || type === "CODE_STRING"}
  class="px-6 size-full flex items-center"
  on:mouseover={handleMouseOver}
  on:focus={handleFocus}
  tabindex="0"
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
