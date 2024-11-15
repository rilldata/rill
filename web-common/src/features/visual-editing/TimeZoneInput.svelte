<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
  import ZoneDisplay from "../dashboards/time-controls/super-pill/components/ZoneDisplay.svelte";
  import SelectionDropdown from "./SelectionDropdown.svelte";

  const defaultSet = new Set(DEFAULT_TIMEZONES);
  const searchableItems = Intl.supportedValuesOf("timeZone");

  export let selectedItems: Set<string>;
  export let keyNotSet: boolean;
  export let onSelectCustomItem: (item: string) => void;
  export let setTimeZones: (timeZones: string[]) => void;

  let hasDefaultsSelected =
    keyNotSet ||
    (defaultSet.size === selectedItems.size &&
      defaultSet.isSubsetOf(selectedItems));

  let selected: 0 | 1 = hasDefaultsSelected ? 0 : 1;

  let selectedProxy = new Set(selectedItems);
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    capitalize={false}
    label="Time zones"
    id="visual-explore-zone"
    hint="Time zones selectable via the dashboard filter bar"
  />
  <FieldSwitcher
    fields={["default", "custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "custom") {
        selected = 1;
        setTimeZones(Array.from(selectedProxy));
      } else if (field === "default") {
        selected = 0;
        setTimeZones(DEFAULT_TIMEZONES);
      }
    }}
  />

  {#if selected === 1}
    <SelectionDropdown
      {searchableItems}
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
        setTimeZones(ranges);
      }}
      let:item
      type="time zones"
    >
      <ZoneDisplay iana={item} />
    </SelectionDropdown>
  {/if}
</div>

<style lang="postcss">
</style>
