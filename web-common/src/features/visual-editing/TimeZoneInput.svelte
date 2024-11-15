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
  export let onSelectMode: (
    mode: "custom" | "default",
    defaults: string[],
  ) => void;
  export let onSelectCustomItem: (item: string) => void;
  export let setTimeZones: (timeZones: string[]) => void;

  $: hasDefaultsSelected =
    keyNotSet ||
    (defaultSet.size === selectedItems.size &&
      defaultSet.isSubsetOf(selectedItems));

  $: mode = hasDefaultsSelected ? "default" : "custom";

  $: selected = mode === "custom" ? 1 : 0;
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    capitalize={false}
    label="Available time zones"
    id="visual-explore-zone"
  />
  <FieldSwitcher
    fields={["Default", "Custom"]}
    {selected}
    onClick={(_, field) => {
      if (field === "Custom") {
        mode = "custom";
        onSelectMode("custom", DEFAULT_TIMEZONES);
      } else if (field === "Default") {
        onSelectMode("default", DEFAULT_TIMEZONES);
        mode = "default";
      }
    }}
  />

  {#if mode === "custom"}
    <SelectionDropdown
      {searchableItems}
      allItems={defaultSet}
      {selectedItems}
      onSelect={onSelectCustomItem}
      setItems={setTimeZones}
      let:item
      type="time zones"
    >
      <ZoneDisplay iana={item} />
    </SelectionDropdown>
  {/if}
</div>

<style lang="postcss">
</style>
