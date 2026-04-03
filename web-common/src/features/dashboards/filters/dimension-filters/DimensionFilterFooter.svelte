<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
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

  const includeExcludeOptions = [
    { value: "include", label: "Include" },
    { value: "exclude", label: "Exclude" },
  ];

  const andOrOptions = [
    {
      value: "or",
      label: "Match any",
      description: "Array contains any selected value",
    },
    {
      value: "and",
      label: "Match all",
      description: "Array contains every selected value",
    },
  ];

  $: includeExcludeValue = excludeMode ? "exclude" : "include";
  $: andOrValue = andMode ? "and" : "or";

  function handleIncludeExcludeChange(value: string) {
    const newExclude = value === "exclude";
    if (newExclude !== excludeMode) {
      onToggleExcludeMode();
    }
  }

  function handleAndOrChange(value: string) {
    const newAnd = value === "and";
    if (newAnd !== andMode) {
      onToggleAndMode?.();
    }
  }
</script>

<footer>
  <div class="flex items-center gap-x-2">
    <Select
      id="include-exclude-mode"
      value={includeExcludeValue}
      options={includeExcludeOptions}
      onChange={handleIncludeExcludeChange}
      size="sm"
      minWidth={82}
    />
    {#if isUnnest && onToggleAndMode}
      <Select
        id="and-or-mode"
        value={andOrValue}
        options={andOrOptions}
        onChange={handleAndOrChange}
        size="sm"
        minWidth={92}
      />
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
