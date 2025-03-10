<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import SettingsSlider from "@rilldata/web-common/components/icons/SettingsSlider.svelte";
  import type {
    ChartSortDirection,
    FieldConfig,
  } from "@rilldata/web-common/features/canvas/components/charts/types";

  export let key: string;
  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;

  $: isDimension = key === "x";
  $: isTemporal = fieldConfig?.type === "temporal";

  let limit = fieldConfig?.limit || 5000;
  let isDropdownOpen = false;

  const sortOptions: { label: string; value: ChartSortDirection }[] = [
    { label: "Ascending", value: "x" },
    { label: "Descending", value: "-x" },
    { label: "Y-axis ascending", value: "y" },
    { label: "Y-axis descending", value: "-y" },
  ];
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <SettingsSlider size="14px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-[280px] p-0">
    <div class="px-3.5 py-2 border-b border-gray-200">
      <span class="text-xs font-medium"
        >{isDimension ? "X-axis" : "Y-axis"} Configuration</span
      >
    </div>
    <div class="px-3.5 pb-1.5">
      <div class="py-1.5 flex items-center justify-between">
        <span class="text-xs">Show axis title</span>
        <Switch
          small
          checked={fieldConfig?.showAxisTitle}
          on:click={() => {
            onChange("showAxisTitle", !fieldConfig?.showAxisTitle);
          }}
        />
      </div>
      {#if isDimension && !isTemporal}
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">Show null values</span>
          <Switch
            small
            checked={fieldConfig?.showNull}
            on:click={() => {
              onChange("showNull", !fieldConfig?.showNull);
            }}
          />
        </div>
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">Sort</span>
          <Select
            size="sm"
            id="sort-select"
            width={180}
            options={sortOptions}
            value={fieldConfig?.sort || "x"}
            on:change={(e) => onChange("sort", e.detail)}
          />
        </div>
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">Limit</span>
          <Input
            size="sm"
            width="72px"
            id="limit-select"
            inputType="number"
            bind:value={limit}
            onBlur={() => {
              onChange("limit", limit);
            }}
            onEnter={() => {
              onChange("limit", limit);
            }}
          />
        </div>
      {/if}
      {#if !isDimension}
        <div class="py-1.5 flex items-center justify-between">
          <span class="text-xs">Zero based origin</span>
          <Switch
            small
            checked={fieldConfig?.zeroBasedOrigin}
            on:click={() => {
              onChange("zeroBasedOrigin", !fieldConfig?.zeroBasedOrigin);
            }}
          />
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
