<script lang="ts">
  import * as DropdownMenu from "../../dropdown-menu";
  import CellInspector from "../../CellInspector.svelte";
  import SearchableMenuContent from "../SearchableMenuContent.svelte";
  export let onSelect: (name: string) => void;
</script>

<DropdownMenu.Root open>
  <DropdownMenu.Trigger>Open filters</DropdownMenu.Trigger>
  <SearchableMenuContent
    selectableGroups={[
      {
        name: "Dimensions",
        items: [
          { name: "country", label: "Country" },
          { name: "device", label: "Device" },
          { name: "publisher", label: "Publisher" },
        ],
      },
    ]}
    selectedItems={[]}
    {onSelect}
  />
</DropdownMenu.Root>

<CellInspector />
