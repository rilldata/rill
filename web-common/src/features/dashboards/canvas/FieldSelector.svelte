<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let selectedItem: string | undefined;
  export let type: "measure" | "dimension";
  export let onSelect: (item: string) => void;
</script>

<InputLabel
  label="X axis"
  id="x-axis"
  hint="Selection of a dimension from the underlying metrics view for inclusion on the dashboard"
/>
<FieldSelectorDropdown {selectedItem} {type} {onSelect} />
