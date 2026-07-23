<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { ChevronDown, ChevronUp, Plus, X } from "lucide-svelte";
  import {
    PIVOT_RULE_COLORS,
    resolveRuleColor,
  } from "./pivot-conditional-formatting";
  import type { PivotFormatRule, PivotFormatRuleOperator } from "./types";

  type Props = {
    // Ordered rule list; first match wins. Precedence is adjusted with the
    // up/down controls on each row.
    rules: PivotFormatRule[];
    onChange: (rules: PivotFormatRule[]) => void;
  };

  let { rules, onChange }: Props = $props();

  const operatorOptions: { value: PivotFormatRuleOperator; label: string }[] = [
    { value: "gt", label: ">" },
    { value: "gte", label: "≥" },
    { value: "lt", label: "<" },
    { value: "lte", label: "≤" },
    { value: "eq", label: "=" },
    { value: "between", label: "between" },
  ];

  function updateRule(index: number, patch: Partial<PivotFormatRule>) {
    const next = rules.map((rule, i) =>
      i === index ? { ...rule, ...patch } : rule,
    );
    // "between" is the only operator using value2; drop it when switching away.
    if (patch.operator && patch.operator !== "between") {
      delete next[index].value2;
    } else if (
      patch.operator === "between" &&
      next[index].value2 === undefined
    ) {
      next[index].value2 = next[index].value;
    }
    onChange(next);
  }

  function moveRule(index: number, delta: -1 | 1) {
    const target = index + delta;
    if (target < 0 || target >= rules.length) return;
    const next = [...rules];
    [next[index], next[target]] = [next[target], next[index]];
    onChange(next);
  }

  function removeRule(index: number) {
    onChange(rules.filter((_, i) => i !== index));
  }

  function addRule() {
    onChange([...rules, { operator: "gt", value: 0, color: "positive" }]);
  }

  function handleValueInput(
    index: number,
    field: "value" | "value2",
    raw: string,
  ) {
    const parsed = Number(raw);
    if (raw === "" || !Number.isFinite(parsed)) return;
    updateRule(index, { [field]: parsed });
  }
</script>

<div class="flex flex-col gap-y-2">
  {#each rules as rule, index (index)}
    <div class="flex items-center gap-x-1">
      <div class="flex flex-col flex-none">
        <button
          type="button"
          class="reorder-btn"
          disabled={index === 0}
          aria-label="Move rule up"
          onclick={() => moveRule(index, -1)}
        >
          <ChevronUp size="10px" />
        </button>
        <button
          type="button"
          class="reorder-btn"
          disabled={index === rules.length - 1}
          aria-label="Move rule down"
          onclick={() => moveRule(index, 1)}
        >
          <ChevronDown size="10px" />
        </button>
      </div>

      <Select
        id="pivot-format-rule-op-{index}"
        value={rule.operator}
        options={operatorOptions}
        onChange={(v) =>
          updateRule(index, { operator: v as PivotFormatRuleOperator })}
        size="sm"
        width={rule.operator === "between" ? 84 : 48}
      />

      <input
        type="number"
        class="value-input"
        value={rule.value}
        aria-label="Rule value"
        onchange={(e) =>
          handleValueInput(index, "value", e.currentTarget.value)}
      />
      {#if rule.operator === "between"}
        <span class="text-fg-secondary text-xs flex-none">and</span>
        <input
          type="number"
          class="value-input"
          value={rule.value2 ?? rule.value}
          aria-label="Rule upper value"
          onchange={(e) =>
            handleValueInput(index, "value2", e.currentTarget.value)}
        />
      {/if}

      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            <button
              {...props}
              type="button"
              class="color-swatch flex-none"
              style:background={resolveRuleColor(rule.color).fill}
              aria-label="Rule color"
            ></button>
          {/snippet}
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="p-2">
          <div class="flex items-center gap-x-1.5">
            {#each PIVOT_RULE_COLORS as color (color.key)}
              <button
                type="button"
                class={["color-swatch", rule.color === color.key && "ring-1"]}
                style:background={color.fill}
                title={color.label}
                aria-label={color.label}
                onclick={() => updateRule(index, { color: color.key })}
              ></button>
            {/each}
            <label
              class="color-swatch custom-color"
              title="Custom color"
              aria-label="Custom color"
            >
              <input
                type="color"
                class="sr-only"
                value={resolveRuleColor(rule.color).fill}
                onchange={(e) =>
                  updateRule(index, { color: e.currentTarget.value })}
              />
            </label>
          </div>
        </DropdownMenu.Content>
      </DropdownMenu.Root>

      <button
        type="button"
        class="remove-btn flex-none"
        aria-label="Remove rule"
        onclick={() => removeRule(index)}
      >
        <X size="12px" />
      </button>
    </div>
  {/each}

  <button type="button" class="add-rule-btn" onclick={addRule}>
    <Plus size="12px" />
    Add rule
  </button>
</div>

<style lang="postcss">
  .reorder-btn {
    @apply flex items-center justify-center h-3 w-4 rounded-sm;
    @apply text-icon-muted hover:text-fg-primary;
  }

  .reorder-btn:disabled {
    @apply opacity-30 pointer-events-none;
  }

  .value-input {
    @apply h-6 w-full min-w-0 rounded-sm border border-gray-300 bg-input px-1.5;
    @apply text-xs text-right;
  }

  .color-swatch {
    @apply h-5 w-5 rounded-sm border border-gray-300 cursor-pointer;
    @apply ring-gray-500 ring-offset-1;
  }

  .custom-color {
    background: conic-gradient(red, yellow, lime, cyan, blue, magenta, red);
  }

  .remove-btn {
    @apply flex items-center justify-center h-5 w-5 rounded-sm;
    @apply text-icon-muted hover:text-fg-primary hover:bg-gray-100;
  }

  .add-rule-btn {
    @apply flex items-center gap-x-1 self-start rounded-sm px-1 py-0.5;
    @apply text-xs text-fg-secondary hover:text-fg-primary hover:bg-gray-100;
  }
</style>
