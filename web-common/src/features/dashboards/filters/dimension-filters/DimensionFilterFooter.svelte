<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let mode: DimensionFilterMode;
  export let excludeMode: boolean;
  export let allSelected: boolean;
  export let disableApplyButton: boolean;
  export let onToggleExcludeMode: (checked: boolean) => void;
  export let onToggleSelectAll: () => void;
  export let onApply: () => void;
</script>

<footer>
  <div class="flex items-center gap-x-1.5">
    <Switch
      checked={excludeMode}
      id="include-exclude"
      small
      onCheckedChange={onToggleExcludeMode}
      label={m.dashboard_include_exclude_toggle()}
    />
    <Label class="font-normal text-xs" for="include-exclude"
      >{m.dashboard_exclude()}</Label
    >
  </div>
  <div class="flex gap-2">
    {#if mode === DimensionFilterMode.Select}
      <Button onClick={onToggleSelectAll} type="tertiary">
        {#if allSelected}
          {m.dashboard_deselect_all()}
        {:else}
          {m.dashboard_select_all()}
        {/if}
      </Button>
    {:else}
      <Button
        onClick={onApply}
        type="primary"
        class="justify-end"
        disabled={disableApplyButton}
      >
        {m.common_apply()}
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
