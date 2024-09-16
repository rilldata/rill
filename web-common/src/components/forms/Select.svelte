<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { InfoIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import * as Select from "@rilldata/web-common/components/select";

  const dispatch = createEventDispatcher();

  export let value: string;
  export let id: string;
  export let label: string;
  export let options: { value: string; label: string }[];
  export let placeholder: string = "";
  export let optional: boolean = false;
  export let tooltip: string = "";
  export let width: number | null = null;

  $: selected = options.find((option) => option.value === value);
</script>

<div class="flex flex-col gap-y-2">
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
    }}
    items={options}
  >
    <Select.Trigger class="px-3 gap-x-2 {width && `w-[${width}px]`}">
      <Select.Value
        {placeholder}
        class="text-[12px] {!selected ? 'text-gray-400' : ''}"
      />
    </Select.Trigger>

    <Select.Content
      sameWidth={false}
      align="start"
      class="max-h-80 overflow-y-auto"
    >
      {#each options as option (option.value)}
        <Select.Item
          value={option.value}
          label={option.label}
          class="text-[12px] "
        >
          {option?.label ?? option.value}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
</div>
