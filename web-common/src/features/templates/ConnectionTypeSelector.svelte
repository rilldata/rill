<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import * as Select from "@rilldata/web-common/components/select";
  import { Cloud, Play, Server, Sparkles } from "lucide-svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

  type ConnectionOption = {
    value: string;
    label: string;
    description?: string;
  };

  export let value: string;
  export let options: ConnectionOption[] = [];
  export let label: string = "";
  export let onChange: (value: string) => void = () => {};

  // Icon and color maps for rich select display.
  // Defaults support ClickHouse and DuckDB deployment types; override via props for other connectors.
  export let iconMap: Record<string, ComponentType<SvelteComponent>> = {
    cloud: Cloud,
    playground: Play,
    "self-managed": Server,
    "self-hosted": Server,
    "rill-managed": Sparkles,
  };

  export let colorMap: Record<string, { bg: string; text: string }> = {
    cloud: { bg: "bg-yellow-100", text: "text-yellow-600" },
    playground: { bg: "bg-green-100", text: "text-green-600" },
    "self-managed": { bg: "bg-purple-100", text: "text-purple-600" },
    "self-hosted": { bg: "bg-purple-100", text: "text-purple-600" },
    "rill-managed": { bg: "bg-blue-100", text: "text-blue-600" },
  };

  function getIcon(optionValue: string): ComponentType<SvelteComponent> {
    return iconMap[optionValue] ?? Server;
  }

  function getColors(optionValue: string): { bg: string; text: string } {
    return (
      colorMap[optionValue] ?? { bg: "bg-gray-100", text: "text-gray-500" }
    );
  }

  $: selectedOption = options.find((opt) => opt.value === value);
  $: SelectedIcon = selectedOption ? getIcon(selectedOption.value) : Server;
  $: selectedColors = selectedOption
    ? getColors(selectedOption.value)
    : { bg: "bg-gray-100", text: "text-gray-500" };

  function handleChange(newValue: string | undefined) {
    if (newValue) {
      value = newValue;
      onChange(newValue);
    }
  }
</script>

<div class="w-full pb-2">
  {#if label}
    <span class="text-sm font-medium text-gray-700 block mb-1">{label}</span>
  {/if}

  <SelectPrimitive.Root
    selected={{ value }}
    onSelectedChange={(s) => handleChange(s?.value)}
  >
    <SelectPrimitive.Trigger
      class="flex h-auto w-full items-center justify-between rounded-[2px] border bg-transparent px-2 py-1.5 text-sm ring-offset-background focus:outline-none focus:border-primary-400"
    >
      {#if selectedOption}
        <div class="flex items-center gap-2">
          <div
            class="flex-shrink-0 w-6 h-6 rounded flex items-center justify-center {selectedColors.bg} {selectedColors.text}"
          >
            <svelte:component this={SelectedIcon} size="14" />
          </div>
          <div class="flex flex-col items-start">
            <span class="text-sm font-medium text-gray-900">
              {selectedOption.label}
            </span>
            {#if selectedOption.description}
              <span class="text-xs text-gray-500">
                {selectedOption.description}
              </span>
            {/if}
          </div>
        </div>
      {:else}
        <span class="text-fg-muted">Select connection type</span>
      {/if}
      <div class="caret transition-transform ml-2">
        <CaretDownIcon size="12px" className="fill-fg-secondary" />
      </div>
    </SelectPrimitive.Trigger>

    <Select.Content class="w-full" sameWidth>
      {#each options as option (option.value)}
        {@const Icon = getIcon(option.value)}
        {@const colors = getColors(option.value)}
        <Select.Item value={option.value} class="py-2">
          <div class="flex items-center gap-3">
            <div
              class="flex-shrink-0 w-7 h-7 rounded-md flex items-center justify-center {colors.bg} {colors.text}"
            >
              <svelte:component this={Icon} size="16" />
            </div>
            <div class="flex flex-col">
              <span class="text-sm font-medium">{option.label}</span>
              {#if option.description}
                <span class="text-xs text-gray-500">{option.description}</span>
              {/if}
            </div>
          </div>
        </Select.Item>
      {/each}
    </Select.Content>
  </SelectPrimitive.Root>
</div>

<style lang="postcss">
  :global(button[aria-expanded="true"] > .caret) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
