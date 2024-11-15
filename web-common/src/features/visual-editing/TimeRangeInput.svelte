<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import {
    PERIOD_TO_DATE_RANGES,
    LATEST_WINDOW_TIME_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
    DEFAULT_TIME_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import SelectionDropdown from "./SelectionDropdown.svelte";

  const ranges = [
    ...Object.keys(LATEST_WINDOW_TIME_RANGES),
    ...Object.keys(PERIOD_TO_DATE_RANGES),
    ...Object.keys(PREVIOUS_COMPLETE_DATE_RANGES),
  ];

  const defaultSet = new Set(ranges);

  export let selectedItems: Set<string>;
  export let keyNotSet: boolean;
  export let onSelectCustomItem: (item: string) => void;
  export let setTimeRanges: (timeRanges: string[]) => void;

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
  <InputLabel
    capitalize={false}
    label="Time ranges"
    id="visual-explore-range"
    hint="Time range shortcuts available via the dashboard filter bar"
  />
  <FieldSwitcher
    fields={["default", "custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "custom") {
        selected = 1;
        setTimeRanges(Array.from(selectedProxy));
      } else if (field === "default") {
        selected = 0;
        setTimeRanges(ranges);
      }
    }}
  />

  {#if selected === 1}
    <SelectionDropdown
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
      setItems={(ranges) => {
        selectedProxy = new Set(ranges);
        setTimeRanges(ranges);
      }}
      let:item
      type="time ranges"
    >
      {DEFAULT_TIME_RANGES[item]?.label ?? item}
    </SelectionDropdown>
  {/if}
</div>

<style lang="postcss">
</style>
