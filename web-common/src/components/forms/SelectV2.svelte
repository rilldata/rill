<script lang="ts">
  import * as Select from "@rilldata/web-common/components/select";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { InfoIcon } from "lucide-svelte";

  export let id: string;
  export let value: string;
  export let options: {
    value: string;
    label?: string;
    disabled?: boolean;
    tooltip?: string;
  }[];
  export let label = "";
  export let optional = false;
  export let tooltip = "";
  export let placeholder = "";
  export let className = "";

  $: selectedOption = options.find((o) => o.value === value);
</script>

{#if label}
  <label for={id} class="text-sm flex items-center gap-x-1">
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
          {tooltip}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}
  </label>
{/if}
<Select.Root
  selected={selectedOption}
  onSelectedChange={(v) => {
    if (!v) return;
    value = v.value ?? "";
  }}
  portal={null}
  items={options}
>
  <Select.Trigger class="flex-none {className}" aria-label="Select a {id}">
    <Select.Value {placeholder} />
  </Select.Trigger>
  <Select.Content class="max-h-64 overflow-y-auto">
    {#each options as { value, label, disabled, tooltip }}
      {#if tooltip}
        <Tooltip.Root portal="body">
          <Tooltip.Trigger>
            <Select.Item {value} {disabled}>
              {label ?? value}
            </Select.Item>
          </Tooltip.Trigger>
          <Tooltip.Content side="right">
            {tooltip}
          </Tooltip.Content>
        </Tooltip.Root>
      {:else}
        <Select.Item {value} {disabled}>
          {label ?? value}
        </Select.Item>
      {/if}
    {/each}
  </Select.Content>
</Select.Root>
