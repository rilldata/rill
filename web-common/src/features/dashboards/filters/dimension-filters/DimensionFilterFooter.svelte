<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";

  export let mode: DimensionFilterMode;
  export let excludeMode: boolean;
  export let allSelected: boolean;
  export let disableApplyButton: boolean;
  export let onToggleExcludeMode: () => void;
  export let onToggleSelectAll: () => void;
  export let onApply: () => void;
</script>

<footer>
  <div class="flex items-center gap-x-1.5">
    <Switch
      checked={excludeMode}
      id="include-exclude"
      small
      on:click={onToggleExcludeMode}
      label="Include exclude toggle"
    />
    <Label class="font-normal text-xs" for="include-exclude">Exclude</Label>
  </div>
  <div class="flex gap-2">
    {#if mode === DimensionFilterMode.Select}
      <Button onClick={onToggleSelectAll} type="plain">
        {#if allSelected}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>
    {:else}
      <Button
        onClick={onApply}
        type="primary"
        class="justify-end"
        disabled={disableApplyButton}
      >
        Apply
      </Button>
    {/if}
  </div>
</footer>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-between;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
