<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { InfoIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import * as Select from "@rilldata/web-common/components/select";

  const dispatch = createEventDispatcher();

  export let value: string = "";
  export let id: string;
  export let label: string = "";
  export let options: { value: string; label: string }[];
  export let placeholder: string = "";
  export let optional: boolean = false;
  export let tooltip: string = "";
  export let width: number | null = null;
  export let selectElement: HTMLButtonElement | undefined = undefined;
  export let full = false;
  export let onChange: (value: string) => void = () => {};
  export let fontSize = 12;
  export let sameWidth = false;
  export let ringFocus = true;

  $: selected = options.find((option) => option.value === value);
</script>

<div class="flex flex-col gap-y-2" class:w-full={full}>
  {#if label?.length}
    <label for={id} class="text-sm flex items-center gap-x-1">
      <span class="text-gray-800 font-medium">
        {label}
      </span>
      {#if optional}
        <span class="text-gray-500">(optional)</span>
      {/if}
      {#if tooltip}
        <Tooltip distance={8}>
          <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
          <TooltipContent
            slot="tooltip-content"
            maxWidth="600px"
            class="whitespace-pre-line"
          >
            {tooltip}
          </TooltipContent>
        </Tooltip>
      {/if}
    </label>
  {/if}

  <Select.Root
    {selected}
    onSelectedChange={(newSelection) => {
      if (!newSelection) return;
      value = newSelection.value;
      dispatch("change", newSelection.value);
      onChange(newSelection.value);
    }}
    items={options}
  >
    <Select.Trigger
      {id}
      bind:el={selectElement}
      class="px-3 gap-x-2 {width && `w-[${width}px]`} {ringFocus
        ? 'focus:ring-2 focus:ring-primary-100'
        : ''}"
    >
      <Select.Value
        {placeholder}
        class="text-[{fontSize}px] {!selected ? 'text-gray-400' : ''}"
      />
    </Select.Trigger>

    <Select.Content {sameWidth} align="start" class="max-h-80 overflow-y-auto">
      {#each options as option (option.value)}
        <Select.Item
          value={option.value}
          label={option.label}
          class="text-[{fontSize}px] "
        >
          {option?.label ?? option.value}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
</div>
