<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";

  export let mode: DimensionFilterMode;
  export let excludeMode: boolean;
  export let andMode: boolean = false;
  export let isUnnest: boolean = false;
  export let allSelected: boolean;
  export let disableApplyButton: boolean;
  export let onToggleExcludeMode: () => void;
  export let onToggleAndMode: (() => void) | undefined = undefined;
  export let onToggleSelectAll: () => void;
  export let onApply: () => void;
</script>

<footer>
  <div class="flex items-center gap-x-3">
    <div class="flex items-center gap-x-1.5">
      <Switch
        checked={excludeMode}
        id="include-exclude"
        small
        onclick={onToggleExcludeMode}
        label="Include exclude toggle"
      />
      <Label class="font-normal text-xs" for="include-exclude">Exclude</Label>
    </div>
    {#if isUnnest && onToggleAndMode}
      <div class="flex items-center gap-x-1.5">
        <Switch
          checked={andMode}
          id="and-or-mode"
          small
          onclick={onToggleAndMode}
          label="Match all toggle"
        />
        <Label class="font-normal text-xs" for="and-or-mode">Match all</Label>
      </div>
    {/if}
  </div>
  <div class="flex gap-2">
    {#if mode === DimensionFilterMode.Select}
      <Button onClick={onToggleSelectAll} type="tertiary">
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
    @apply border-t;
    @apply bg-popover-footer;
    @apply flex flex-row flex-none items-center justify-between;
    @apply gap-x-2 p-2 px-3.5;
  }
</style>
