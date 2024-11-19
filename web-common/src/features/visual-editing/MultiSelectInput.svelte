<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import SelectionDropdown from "./SelectionDropdown.svelte";

  export let defaultItems: string[];
  export let searchableItems: string[] | undefined = undefined;
  export let selectedItems: Set<string>;
  export let keyNotSet: boolean;
  export let label: string;
  export let id: string;
  export let hint: string;
  export let onSelectCustomItem: (item: string) => void;
  export let setItems: (timeRanges: string[]) => void;

  const defaultSet = new Set(defaultItems);

  let hasDefaultsSelected =
    keyNotSet ||
    (defaultSet.size === selectedItems.size &&
      defaultSet.isSubsetOf(selectedItems));

  let selected: 0 | 1 = hasDefaultsSelected ? 0 : 1;

  let selectedProxy = new Set(selectedItems);

  $: if (keyNotSet) {
    selected = 0;
  }
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel capitalize={false} {label} {id} {hint} />
  <FieldSwitcher
    fields={["default", "custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "custom") {
        selected = 1;
        setItems(selectedProxy.size ? Array.from(selectedProxy) : defaultItems);
      } else if (field === "default") {
        selected = 0;
        setItems(defaultItems);
      }
    }}
  />

  {#if selected === 1}
    <SelectionDropdown
      {searchableItems}
      {id}
      allItems={defaultSet}
      {selectedItems}
      onSelect={(item) => {
        const deleted = selectedProxy.delete(item);
        if (!deleted) {
          selectedProxy.add(item);
        }

        selectedProxy = selectedProxy;

        onSelectCustomItem(item);
      }}
      setItems={(items) => {
        selectedProxy = new Set(items);
        setItems(items);
      }}
      let:item
      type={label.toLowerCase()}
    >
      <slot {item} />
    </SelectionDropdown>
  {/if}
</div>
