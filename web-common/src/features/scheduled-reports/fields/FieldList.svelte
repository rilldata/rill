<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem.ts";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import ChipDragList from "@rilldata/web-common/features/canvas/inspector/ChipDragList.svelte";
  import type { FieldType } from "@rilldata/web-common/features/canvas/inspector/types.ts";
  import { PlusIcon } from "lucide-svelte";

  export let fields: string[];
  export let allowedFields: string[];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let label: string;
  export let disableDragDrop: boolean = false;
  export let onUpdate: (newFields: string[]) => void;

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
      onUpdate([...fields, name]);
    } else {
      onUpdate([...fields.slice(0, index), ...fields.slice(index + 1)]);
    }
    open = false;
  }

  function handleRemove(item: string) {
    const temp = [...fields];
    const index = temp.indexOf(item);
    if (index !== -1) {
      temp.splice(index, 1);
      onUpdate(temp);
    }
  }
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel {label} id={label} capitalize={false} />

  <div
    class="flex flex-row items-center min-h-7"
    aria-label="{label} field list"
  >
    {#if !fields.length}
      <slot name="empty-fields" />
    {:else if disableDragDrop}
      <div class="flex flex-row flex-wrap gap-1">
        {#each fields as field (field)}
          <Chip
            removable
            fullWidth
            type={displayMap[field]?.type ?? "dimension"}
            onRemove={() => handleRemove(field)}
          >
            <span class="font-bold truncate" slot="body">
              {displayMap[field]?.label || field}
            </span>
          </Chip>
        {/each}
      </div>
    {:else}
      <ChipDragList
        items={fields}
        {onUpdate}
        orientation="horizontal"
        {displayMap}
      />
    {/if}

    <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
      <DropdownMenu.Trigger asChild let:builder>
        <Button
          label={`Add ${label} fields`}
          active={open}
          builders={[builder]}
          class="w-[34px] ml-2 border border-dashed border-slate-300"
          compact
          rounded
        >
          <PlusIcon size="14px" strokeWidth={3} />
        </Button>
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
