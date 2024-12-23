<script>
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let label: string;
  export let id: string;
  export let metricName: string;
  export let selectedItem: string | undefined;
  export let type: "measure" | "dimension";
  export let searchableItems: string[] | undefined = undefined;
  export let onSelect: (item: string) => void;
  export let multi = false;
  export let selectedItems: Set<string> | undefined = undefined;
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <InputLabel
    small
    {label}
    {id}
    hint="Selection of a {type} from the underlying metrics view for inclusion on the dashboard"
  />

  <FieldSelectorDropdown {metricName} {type} {selectedItem} {onSelect} />
</div>
