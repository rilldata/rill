<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";

  export let isDimension: boolean;
  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;

  let isDropdownOpen = false;
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-[250px]">
    <div class="px-2 py-1.5 flex items-center justify-between">
      <span class="text-xs">Show axis title</span>
      <Switch
        small
        checked={fieldConfig?.showAxisTitle}
        on:click={() => {
          onChange("showAxisTitle", !fieldConfig?.showAxisTitle);
        }}
      />
    </div>
    {#if !isDimension}
      <div class="px-2 py-1.5 flex items-center justify-between">
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
  </DropdownMenu.Content>
</DropdownMenu.Root>
