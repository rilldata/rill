<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FieldSelectorDropdown from "@rilldata/web-common/features/canvas/inspector/fields/FieldSelectorDropdown.svelte";

  export let selectedItems: string[];
  export let onMultiSelect: (items: string[]) => void;
  export let allowMultiSelect = true;
  export let open = false;
</script>

<FieldSelectorDropdown
  canvasName="test"
  metricName="ads"
  types={["measure"]}
  {selectedItems}
  {onMultiSelect}
  {allowMultiSelect}
  bind:open
>
  <svelte:fragment slot="trigger">
    <DropdownMenu.Trigger>Choose fields</DropdownMenu.Trigger>
  </svelte:fragment>
</FieldSelectorDropdown>
