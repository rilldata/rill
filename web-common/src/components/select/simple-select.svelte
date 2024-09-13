<script lang="ts">
  import * as Select from "@rilldata/web-common/components/select";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { getAttrs, builderActions } from "bits-ui";

  export let id: string;
  export let value: string;
  export let options: {
    label: string;
    value: string;
    disabled?: boolean;
    tooltip?: string;
  }[];
  export let className = "";

  $: selectedOption = options.find((o) => o.value === value);
</script>

<Select.Root portal={null}>
  <Select.Trigger class="flex-none w-32" aria-label="Select a {id}">
    <Select.Value placeholder={selectedOption?.label ?? `Select a ${id}`} />
  </Select.Trigger>
  <Select.Content class="max-h-64 overflow-y-auto {className}">
    {#each options as { value, label, disabled, tooltip }}
      {#if tooltip}
        <Tooltip.Root portal="body">
          <Tooltip.Trigger>
            <Select.Item {value} {disabled}>
              {label}
            </Select.Item>
          </Tooltip.Trigger>
          <Tooltip.Content side="right">
            {tooltip}
          </Tooltip.Content>
        </Tooltip.Root>
      {:else}
        <Select.Item {value} {disabled}>
          {label}
        </Select.Item>
      {/if}
    {/each}
  </Select.Content>
</Select.Root>
