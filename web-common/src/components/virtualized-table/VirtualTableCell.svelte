<script lang="ts">
  import FormattedDataType from "../../components/data-types/FormattedDataType.svelte";
  import { cellInspectorStore } from "../../features/dashboards/stores/cellInspectorStore";

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

  let isOpen = false;

  function handleMouseOver(e) {
    if (value !== undefined && value !== null) {
      // Always update the value in the store, but don't change visibility
      cellInspectorStore.updateValue(value.toString());
    }
  }

  function handleFocus() {
    if (value !== undefined && value !== null) {
      // Always update the value in the store, but don't change visibility
      cellInspectorStore.updateValue(value.toString());
    }
  }

  function handleKeyDown(e) {
    if (e.code === "Space" || e.code === "Enter") {
      e.preventDefault();
      e.stopPropagation();
      isOpen = !isOpen;

      if (isOpen && value !== undefined && value !== null) {
        // Only open the inspector when spacebar is pressed
        cellInspectorStore.open(value.toString());
      } else {
        // Close the inspector
        cellInspectorStore.close();
      }
    } else if (e.key === "Escape" && isOpen) {
      e.preventDefault();
      e.stopPropagation();
      isOpen = false;
      cellInspectorStore.close();
    }
  }
</script>

<div
  class:sorted
  class:selected
  class:!justify-start={type === "VARCHAR" || type === "CODE_STRING"}
  class=" px-6 size-full flex items-center"
  on:mouseover={handleMouseOver}
  on:focus={handleFocus}
  on:keydown={handleKeyDown}
  role="cell"
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
