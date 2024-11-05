<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import SelectionDropdown from "./SelectionDropdown.svelte";

  export let type: "measures" | "dimensions";
  export let mode: "all" | "subset" | "expression";
  export let expression: string = "";
  export let items: (MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2)[];
  export let selectedItems: Set<string> | undefined;
  export let onSelectAll: () => void;
  export let onSelectSubset: () => void;
  export let onSelectSubsetItem: (item: string) => void;
  export let onSelectExpression: () => void;
  export let onExpressionBlur: (value: string) => void;

  $: selected = mode === "all" ? 0 : mode === "subset" ? 1 : 2;
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel label={type} id="visual-explore-{type}" />
  <FieldSwitcher
    fields={["All", "Subset", "Expression"]}
    {selected}
    onClick={async (i, field) => {
      if (field === "All") {
        onSelectAll();
      } else if (field === "Subset") {
        onSelectSubset();
      } else {
        onSelectExpression();
      }
    }}
  />

  {#if mode === "expression"}
    <Input
      textClass="text-sm"
      multiline
      bind:value={expression}
      onBlur={() => {
        onExpressionBlur(expression);
      }}
      onEnter={() => {
        onExpressionBlur(expression);
      }}
    />
  {:else if mode === "subset"}
    <SelectionDropdown
      allItems={new Set(items.map((m) => m.name ?? ""))}
      selectedItems={new Set(selectedItems)}
      onSelect={onSelectSubsetItem}
      onToggleSelectAll={() => {}}
    />
  {/if}
</div>
