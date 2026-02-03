<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import {
    CriteriaGroupOptions,
    CriteriaOperationOptions,
  } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import {
    MeasureFilterBaseTypeOptions,
    MeasureFilterComparisonTypeOptions,
    MeasureFilterPercentOfTotalOption,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import { Trash2Icon } from "lucide-svelte";

  export let criteria: MeasureFilterEntry[];
  export let criteriaOperation: V1Operation;
  export let hasTimeComparison: boolean;
  export let measure: string;
  export let metricsViewSpec: V1MetricsViewSpec;
  export let onAddCriteria: () => void;
  export let onRemoveCriteria: (index: number) => void;
  export let onUpdateCriteria: (
    index: number,
    field: string,
    value: unknown,
  ) => void;
  export let onOperationChange: (operation: V1Operation) => void;

  $: selectedMeasure = metricsViewSpec.measures?.find((m) => m.name === measure);

  $: typeOptions = (() => {
    const options = [...MeasureFilterBaseTypeOptions];

    if (hasTimeComparison) {
      options.push(
        ...MeasureFilterComparisonTypeOptions.map((o) => ({
          ...o,
          label: o.shortLabel,
        })),
      );
    } else {
      options.push(
        ...MeasureFilterComparisonTypeOptions.map((o) => ({
          ...o,
          label: o.shortLabel,
          tooltip: "Available when comparing time periods.",
          disabled: true,
        })),
      );
    }

    if (selectedMeasure?.validPercentOfTotal) {
      options.push(MeasureFilterPercentOfTotalOption);
    } else {
      options.push({
        ...MeasureFilterPercentOfTotalOption,
        tooltip: "Measure does not support percent-of-total.",
        disabled: true,
      });
    }

    return options;
  })();

  $: measureLabel =
    selectedMeasure?.displayName || selectedMeasure?.name || measure;
</script>

<div class="flex flex-col gap-y-3">
  {#each criteria as criterion, index (index)}
    {#if index > 0}
      <div class="flex items-center justify-center py-1">
        <Select
          bind:value={criteriaOperation}
          id="criteria-operation-{index}"
          label=""
          ariaLabel="Criteria group operation"
          options={CriteriaGroupOptions}
          onChange={(value) => onOperationChange(value)}
          width={80}
        />
      </div>
    {/if}

    <div
      class="flex flex-col gap-y-2 p-3 bg-surface-subtle rounded-md border border-surface-border"
    >
      <div class="flex items-center justify-between">
        <span class="text-sm font-medium text-fg-secondary">
          Criterion {index + 1}
        </span>
        <button
          type="button"
          class="text-fg-muted hover:text-fg-primary transition-colors"
          on:click={() => onRemoveCriteria(index)}
          aria-label="Remove criterion {index + 1}"
        >
          <Trash2Icon size="16px" />
        </button>
      </div>

      <div class="flex flex-col gap-y-2">
        <div class="text-sm text-fg-secondary truncate" title={measureLabel}>
          {measureLabel}
        </div>

        <div class="flex gap-x-2">
          <Select
            value={criterion.type}
            id="criteria-type-{index}"
            label=""
            ariaLabel="Criteria type"
            options={typeOptions}
            onChange={(value) => onUpdateCriteria(index, "type", value)}
            width={140}
          />
          <Select
            value={criterion.operation}
            id="criteria-operation-{index}"
            label=""
            ariaLabel="Criteria operator"
            options={CriteriaOperationOptions}
            onChange={(value) => onUpdateCriteria(index, "operation", value)}
            width={60}
          />
          <Input
            value={criterion.value1}
            id="criteria-value-{index}"
            title="Criteria value"
            onInput={(value, _e) => onUpdateCriteria(index, "value1", value)}
            placeholder="0"
            width="auto"
          />
        </div>
      </div>
    </div>
  {/each}

  <Button type="tertiary" onClick={onAddCriteria}>+ Add Criterion</Button>

  {#if criteria.length === 0}
    <p class="text-sm text-fg-muted">
      Add at least one criterion to trigger this alert.
    </p>
  {/if}
</div>
