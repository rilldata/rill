<script lang="ts">
  import { SelectSeparator } from "@rilldata/web-common/components/select";
  import * as Select from "@rilldata/web-common/components/select";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { InfoIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import DataTypeIcon from "../data-types/DataTypeIcon.svelte";
  import Search from "../search/Search.svelte";

  const dispatch = createEventDispatcher();

  export let value: string = "";
  export let id: string;
  export let label: string = "";
  export let ariaLabel: string = ""; // Fallback aria-label attribute if label is empty
  export let lockTooltip: string = "";
  export let size: "sm" | "md" | "lg" | "xl" = "lg";
  export let options: {
    value: string;
    label: string;
    description?: string;
    type?: string;
    disabled?: boolean;
    tooltip?: string;
  }[];
  export let optionsLoading: boolean = false;
  export let onAddNew: (() => void) | null = null;
  export let addNewLabel: string | null = null;
  export let placeholder: string = "";
  export let optional: boolean = false;
  export let tooltip: string = "";
  export let width: number | null = null;
  export let minWidth: number | null = null;
  export let dropdownWidth: string | null = null;
  export let disabled = false;
  export let selectElement: HTMLButtonElement | undefined = undefined;
  export let full = false;
  export let fontSize = 12;
  export let sameWidth = false;
  export let ringFocus = true;
  export let truncate = false;
  export let enableSearch = false;
  export let lockable = false;
  export let forcedTriggerStyle = "";
  export let onChange: (value: string) => void = () => {};

  let searchText = "";

  const HeightBySize = {
    sm: "h-6",
    md: "h-7",
    lg: "",
  };

  $: selected = options.find((option) => option.value === value);
  $: filteredOptions = enableSearch
    ? options.filter((option) =>
        option.label.toLowerCase().includes(searchText.toLowerCase()),
      )
    : options;

  let open = false;
</script>

<div class="flex flex-col gap-y-2 max-w-full" class:w-full={full}>
  {#if label?.length}
    <label
      for={id}
      class="{size === 'sm' ? 'text-xs' : 'text-sm'} flex items-center gap-x-1"
    >
      <span class="text-gray-800 font-medium">
        {label}
      </span>
      {#if optional}
        <span class="text-gray-500">(optional)</span>
      {/if}
      {#if tooltip}
        <Tooltip.Root portal="body">
          <Tooltip.Trigger>
            <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
          </Tooltip.Trigger>
          <Tooltip.Content side="right">
            {#each tooltip.split(/\n/gm) as line (line)}
              <div>{line}</div>
            {/each}
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}
    </label>
  {/if}

  <Select.Root
    bind:open
    {disabled}
    {selected}
    onSelectedChange={(newSelection) => {
      if (!newSelection) return;
      value = newSelection.value;
      dispatch("change", newSelection.value);
      onChange(newSelection.value);
    }}
    onOpenChange={(isOpen) => {
      if (!isOpen) {
        searchText = "";
      }
    }}
    items={options}
  >
    <Select.Trigger
      {id}
      {disabled}
      {lockable}
      {lockTooltip}
      bind:el={selectElement}
      class="flex px-3 gap-x-2 max-w-full {HeightBySize[size]} {width &&
        `w-[${width}px]`} {minWidth && `min-w-[${minWidth}px]`} {ringFocus &&
        'focus:ring-2 focus:ring-primary-100'} {truncate
        ? 'break-all overflow-hidden'
        : ''} {forcedTriggerStyle}"
      aria-label={label || ariaLabel}
    >
      <Select.Value
        {placeholder}
        class="text-[{fontSize}px] {!selected
          ? 'text-gray-400'
          : ''} w-full  text-left"
      />
    </Select.Trigger>

    <Select.Content
      {sameWidth}
      align="start"
      class="max-h-80 overflow-y-auto {dropdownWidth ? dropdownWidth : ''}"
      strategy="fixed"
    >
      {#if enableSearch}
        <div class="px-2 py-1.5">
          <Search bind:value={searchText} showBorderOnFocus={false} />
        </div>
      {/if}
      {#if optionsLoading}
        <div class="flex flex-row items-center ml-5 h-10 w-full">
          <div class="m-auto w-10">
            <Spinner size="18px" status={EntityStatus.Running} />
          </div>
        </div>
      {:else}
        {#each filteredOptions as { type, value, label, description, disabled, tooltip } (value)}
          <Select.Item
            {value}
            {label}
            {description}
            {disabled}
            class="text-[{fontSize}px] gap-x-2 items-start"
          >
            {#if tooltip}
              <Tooltip.Root portal="body">
                <Tooltip.Trigger class="select-tooltip cursor-default">
                  {#if type}
                    <DataTypeIcon {type} />
                  {/if}
                  {label ?? value}
                </Tooltip.Trigger>
                <Tooltip.Content side="right" sideOffset={8}>
                  {tooltip}
                </Tooltip.Content>
              </Tooltip.Root>
            {:else}
              {#if type}
                <DataTypeIcon {type} />
              {/if}
              {label ?? value}
            {/if}
          </Select.Item>
        {:else}
          <div class="px-2.5 py-1.5 text-gray-600">No results found</div>
        {/each}
        {#if onAddNew}
          <SelectSeparator />
          <Select.Item
            value="__rill_add_option__"
            on:click={(e) => {
              e.stopPropagation();
              e.preventDefault();
              open = false;
              onAddNew();
            }}
            class="text-[{fontSize}px]"
          >
            {addNewLabel ?? "+ Add"}
          </Select.Item>
        {/if}
      {/if}
    </Select.Content>
  </Select.Root>
</div>
