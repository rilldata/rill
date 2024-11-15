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
  export let onSelectSubsetItem: (item: string) => void;
  export let onSelectExpression: () => void;
  export let onExpressionBlur: (value: string) => void;
  export let setItems: (items: string[]) => void;

  let selectedProxy = new Set(selectedItems);

  $: selected = mode === "all" ? 0 : mode === "subset" ? 1 : 2;

  function isString(value: unknown): value is string {
    return typeof value === "string";
  }
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    label={type}
    id="visual-explore-{type}"
    hint="Selection of {type} from the underlying metrics view for inclusion on the dashboard"
  />
  <FieldSwitcher
    fields={["all", "subset", "expression"]}
    {selected}
    onClick={async (i, field) => {
      if (field === "all") {
        onSelectAll();
      } else if (field === "subset") {
        if (selectedProxy.size) {
          setItems(Array.from(selectedProxy));
        } else {
          setItems(items.map(({ name }) => name).filter(isString));
        }
      } else if (field === "expression") {
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
      onSelect={(item) => {
        const deleted = selectedProxy.delete(item);
        if (!deleted) {
          selectedProxy.add(item);
        }

        selectedProxy = selectedProxy;

        onSelectSubsetItem(item);
      }}
      setItems={(items) => {
        selectedProxy = new Set(items);
        setItems(items);
      }}
    />
  {/if}
</div>
