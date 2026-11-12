<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
  import { useDimensionSearch } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-values";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ChevronDown } from "lucide-svelte";
  import {
    controlLabel,
    controlOptionLabel,
    controlSummary,
  } from "./control-labels";
  import {
    ControlKind,
    defaultControlValues,
    type AppliedControl,
  } from "./controls";

  // One viewer control: a button group for a short single choice, otherwise a dropdown.
  let {
    control,
    fieldLabel,
    metricsViewNames = [],
    timeStart,
    timeEnd,
    onChange,
  }: {
    control: AppliedControl;
    // Display name of a measure or dimension.
    fieldLabel: (name: string) => string;
    // Metrics views to read a filter control's values from.
    metricsViewNames?: string[];
    timeStart?: string;
    timeEnd?: string;
    // Null resets the control to its default.
    onChange: (values: string[] | null) => void;
  } = $props();

  const client = useRuntimeClient();

  let open = $state(false);
  let searchText = $state("");

  let spec = $derived(control.control);
  let kind = $derived(spec.kind);
  let isFilter = $derived(kind === ControlKind.Filter);
  let multiple = $derived(!!spec.multiple);
  let values = $derived(control.values ?? []);

  let label = $derived(controlLabel(spec, fieldLabel));

  // A filter without fixed options offers the dimension's values, searched as the viewer types.
  let queriesValues = $derived(isFilter && !spec.options?.length);
  let dimension = $derived(
    spec.fields?.[metricsViewNames[0] ?? ""] ?? spec.target ?? "",
  );
  let searchQuery = $derived(
    useDimensionSearch(
      client,
      queriesValues ? metricsViewNames : [],
      dimension,
      {
        mode: DimensionFilterMode.Select,
        searchText,
        values,
        timeStart,
        timeEnd,
        enabled: queriesValues && open,
      },
    ),
  );
  let options = $derived(
    queriesValues ? ($searchQuery.data ?? values) : (spec.options ?? []),
  );

  function optionLabel(value: string): string {
    return controlOptionLabel(spec, value, fieldLabel);
  }

  let isDefault = $derived.by(() => {
    const def = defaultControlValues(spec);
    if (control.values === null || def === null) return control.values === def;
    return (
      def.length === control.values.length &&
      def.every((v, i) => v === control.values?.[i])
    );
  });

  let summary = $derived(controlSummary(spec, values, fieldLabel));

  let asButtons = $derived(
    !multiple &&
      !queriesValues &&
      !isFilter &&
      (spec.display === "buttons" ||
        (spec.display !== "dropdown" && options.length <= 4)),
  );

  function toggle(value: string) {
    if (!multiple) {
      onChange([value]);
      open = false;
      return;
    }
    onChange(
      values.includes(value)
        ? values.filter((v) => v !== value)
        : [...values, value],
    );
  }
</script>

<div class="control" class:changed={!isDefault}>
  {#if label}
    <span class="control-label">{label}</span>
  {/if}

  {#if asButtons}
    <div class="buttons" role="group" aria-label={label || spec.name}>
      {#each options as option (option)}
        <button
          type="button"
          class:selected={values.includes(option)}
          aria-pressed={values.includes(option)}
          onclick={() => onChange([option])}
        >
          {optionLabel(option)}
        </button>
      {/each}
    </div>
  {:else}
    <DropdownMenu.Root bind:open>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <button
            {...props}
            type="button"
            class="trigger"
            aria-label={label || spec.name}
          >
            <span class="truncate">{summary}</span>
            <ChevronDown size="12px" />
          </button>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="start"
        class="min-w-48 max-h-80 overflow-y-auto"
      >
        {#if queriesValues}
          <input
            class="search"
            type="text"
            placeholder={m.canvas_control_search()}
            aria-label={m.canvas_control_search()}
            bind:value={searchText}
            onkeydown={(e) => e.stopPropagation()}
          />
        {/if}
        {#each options as option (option)}
          <DropdownMenu.CheckboxItem
            closeOnSelect={!multiple}
            checked={values.includes(option)}
            onCheckedChange={() => toggle(option)}
          >
            {optionLabel(option)}
          </DropdownMenu.CheckboxItem>
        {:else}
          <div class="px-2 py-1 text-xs text-fg-secondary">
            {m.canvas_control_no_values()}
          </div>
        {/each}
        {#if !isDefault}
          <DropdownMenu.Separator />
          <DropdownMenu.Item onclick={() => onChange(null)}>
            {m.canvas_control_reset()}
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>

<style lang="postcss">
  .control {
    @apply flex items-center gap-x-1.5 text-xs min-w-0;
  }

  .control-label {
    @apply text-fg-secondary font-medium whitespace-nowrap;
  }

  .trigger {
    @apply flex items-center gap-x-1 h-6 px-2 max-w-48;
    @apply border rounded-sm bg-surface-background text-fg-primary;
  }

  .trigger:hover {
    @apply bg-surface-hover;
  }

  .changed .trigger {
    @apply border-primary-400 text-primary-700;
  }

  .buttons {
    @apply flex h-6 border rounded-sm overflow-hidden;
  }

  .buttons button {
    @apply px-2 bg-surface-background text-fg-primary whitespace-nowrap;
  }

  .buttons button + button {
    @apply border-l;
  }

  .buttons button.selected {
    @apply bg-primary-50 text-primary-700 font-medium;
  }

  .search {
    @apply w-full h-7 px-2 mb-1 text-xs border rounded-sm bg-surface-background;
  }
</style>
