<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem.ts";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import ChipDragList from "@rilldata/web-common/features/canvas/inspector/ChipDragList.svelte";
  import type { FieldType } from "@rilldata/web-common/features/canvas/inspector/types.ts";
  import { PlusIcon } from "lucide-svelte";

  export let fields: string[];
  export let allowedFields: string[];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let label: string;

  let open = false;
  let searchValue = "";

  $: allowedMeasures = allowedFields.filter(
    (field) => displayMap[field]?.type === "measure",
  );
  $: allowedTimeFields = allowedFields.filter(
    (field) => displayMap[field]?.type === "time",
  );
  $: allowedDimensions = allowedFields.filter(
    (field) => displayMap[field]?.type === "dimension",
  );

  $: selectableGroups = [
    ...(allowedMeasures.length
      ? [
          <SearchableFilterSelectableGroup>{
            name: "measure",
            label: "MEASURES",
            items: allowedMeasures.map((item) => ({
              name: item,
              label: displayMap[item].label,
            })),
          },
        ]
      : []),
    ...(allowedTimeFields.length
      ? [
          <SearchableFilterSelectableGroup>{
            name: "time",
            label: "TIME",
            items: allowedTimeFields.map((item) => ({
              name: item,
              label: displayMap[item].label,
            })),
          },
        ]
      : []),
    ...(allowedDimensions.length
      ? [
          <SearchableFilterSelectableGroup>{
            name: "dimension",
            label: "DIMENSIONS",
            items: allowedDimensions.map((item) => ({
              name: item,
              label: displayMap[item].label,
            })),
          },
        ]
      : []),
  ];
  $: selectedItems = [
    fields.filter((field) => displayMap[field]?.type === "measure"),
    fields.filter((field) => displayMap[field]?.type === "time"),
    fields.filter((field) => displayMap[field]?.type === "dimension"),
  ];

  function handleSelect(name: string) {
    const index = fields.indexOf(name);
    if (index === -1) {
      fields = [...fields, name];
    } else {
      fields = [...fields.slice(0, index), ...fields.slice(index + 1)];
    }
    open = false;
  }
</script>

<div class="flex flex-col gap-y-1">
  <div>{label}</div>

  <div class="flex flex-row items-center">
    <ChipDragList
      items={fields}
      onUpdate={(newFields) => (fields = newFields)}
      orientation="horizontal"
      {displayMap}
    />

    <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
      <DropdownMenu.Trigger asChild let:builder>
        <button
          aria-label={`Add ${label} fields`}
          use:builder.action
          {...builder}
          class="text-sm px-2 h-6"
        >
          <PlusIcon size="14px" />
        </button>
      </DropdownMenu.Trigger>

      <SearchableMenuContent
        {selectableGroups}
        {selectedItems}
        searchText={searchValue}
        allowSelectAll={false}
        onSelect={handleSelect}
      />
    </DropdownMenu.Root>
  </div>
</div>
