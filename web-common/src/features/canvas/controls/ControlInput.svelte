<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
  import { useDimensionSearch } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-values";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { getQueryServiceMetricsViewAggregationQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ChevronDown } from "lucide-svelte";
  import {
    controlLabel,
    controlOptionLabel,
    controlSummary,
  } from "./control-labels";
  import {
    ControlKind,
    controlValueKey,
    type ControlValue,
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
    onChange: (values: ControlValue[] | null) => void;
  } = $props();

  const client = useRuntimeClient();

  let open = $state(false);
  let searchText = $state("");

  let spec = $derived(control.control);
  let kind = $derived(spec.kind);
  let isFilter = $derived(kind === ControlKind.Filter);
  let multiple = $derived(!!spec.multiple);
  // On its default, a slot control shows the fields the widget's YAML sets there.
  let values = $derived(control.values ?? control.shownValues ?? []);
  // Only a filter can be cleared; the other kinds always keep a value.
  let keepsOne = $derived(multiple && !isFilter);

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
  let options = $derived.by(() => {
    const options: ControlValue[] = queriesValues
      ? ($searchQuery.data ?? values)
      : (spec.options ?? []);
    // Search results include the selected values, which can repeat a value as text.
    const unique = new Map<string, ControlValue>();
    for (const value of options) {
      if (!unique.has(controlValueKey(value)))
        unique.set(controlValueKey(value), value);
    }
    return [...unique.values()];
  });

  function optionLabel(value: ControlValue): string {
    return controlOptionLabel(spec, value, fieldLabel);
  }

  let isDefault = $derived.by(() => {
    const def = defaultControlValues(spec);
    const current = control.values;
    if (current === null || def === null) return current === def;
    return (
      def.length === current.length &&
      def.every((v, i) => controlValueKey(v) === controlValueKey(current[i]))
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

  function selected(value: ControlValue): boolean {
    return values.some((v) => controlValueKey(v) === controlValueKey(value));
  }

  function retry() {
    void Promise.all(
      metricsViewNames.map((metricsView) =>
        queryClient.refetchQueries({
          queryKey: getQueryServiceMetricsViewAggregationQueryKey(
            client.instanceId,
            {
              metricsView,
              dimensions: [{ name: dimension }],
            },
          ),
        }),
      ),
    ).catch(console.error);
  }

  function toggle(value: ControlValue) {
    if (!multiple) {
      onChange(isFilter && selected(value) ? [] : [value]);
      open = false;
      return;
    }
    // A pick starts from what the widget shows, as far as the options allow.
    const current = isFilter
      ? values
      : values.filter((v) =>
          options.some((o) => controlValueKey(o) === controlValueKey(v)),
        );
    const next = selected(value)
      ? current.filter((v) => controlValueKey(v) !== controlValueKey(value))
      : [...current, value];
    if (keepsOne && next.length === 0) return;
    // Picking the widget's own fields again goes back to the default.
    const shown = control.shownValues;
    const backToShown =
      !!shown &&
      shown.length === next.length &&
      shown.every((v, i) => controlValueKey(v) === controlValueKey(next[i]));
    onChange(backToShown ? null : next);
  }
</script>

<div class="control" class:changed={!isDefault}>
  {#if label}
    <span class="control-label">{label}</span>
  {/if}

  {#if asButtons}
    <div class="buttons" role="group" aria-label={label || spec.name}>
      {#each options as option (controlValueKey(option))}
        <button
          type="button"
          class:selected={selected(option)}
          aria-pressed={selected(option)}
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
        {#if queriesValues && $searchQuery.isLoading}
          <div role="status" class="px-2 py-1 text-xs text-fg-secondary">
            {m.common_loading()}
          </div>
        {:else if queriesValues && $searchQuery.error}
          <div role="alert" class="px-2 py-1 text-xs text-fg-secondary">
            {m.canvas_control_load_error()}
          </div>
          <DropdownMenu.Item onclick={retry} closeOnSelect={false}
            >{m.common_retry()}</DropdownMenu.Item
          >
        {:else}
          {#each options as option (controlValueKey(option))}
            <DropdownMenu.CheckboxItem
              closeOnSelect={!multiple}
              checked={selected(option)}
              disabled={keepsOne && selected(option) && values.length === 1}
              onCheckedChange={() => toggle(option)}
            >
              {optionLabel(option)}
            </DropdownMenu.CheckboxItem>
          {:else}
            <div class="px-2 py-1 text-xs text-fg-secondary">
              {m.canvas_control_no_values()}
            </div>
          {/each}
        {/if}
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
