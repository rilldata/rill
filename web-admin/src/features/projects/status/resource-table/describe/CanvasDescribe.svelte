<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { V1Canvas } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";

  export let canvas: V1Canvas;

  const dispatch = createEventDispatcher<{
    "view-component": { componentName: string };
  }>();

  $: spec = canvas?.spec;
  $: state = canvas?.state;
  $: rows = spec?.rows ?? [];
  $: variables = spec?.variables ?? [];

  // Count total components across all rows
  $: componentCount = rows.reduce(
    (sum, row) => sum + (row.items?.length ?? 0),
    0,
  );
</script>

<div class="flex flex-col gap-y-3">
  <!-- General -->
  <DescribeSection title="General">
    {#if spec?.theme}
      <DescribeRow label="Theme" value={spec.theme} />
    {/if}
    {#if spec?.banner}
      <DescribeRow label="Banner" value={spec.banner} mono={false} />
    {/if}
    <DescribeRow
      label="Filters enabled"
      value={String(!!spec?.filtersEnabled)}
    />
    <DescribeRow
      label="Allow custom time range"
      value={String(!!spec?.allowCustomTimeRange)}
    />
    {#if state?.dataRefreshedOn}
      <DescribeRow
        label="Data refreshed on"
        value={new Date(state.dataRefreshedOn).toLocaleString()}
        mono={false}
      />
    {/if}
  </DescribeSection>

  <!-- Layout -->
  <DescribeSection title="Layout">
    <DescribeRow label="Rows" value={rows.length} mono={false} />
    <DescribeRow label="Components" value={componentCount} mono={false} />
    {#if spec?.maxWidth}
      <DescribeRow label="Max width" value="{spec.maxWidth}px" mono={false} />
    {/if}
    {#if spec?.gapX}
      <DescribeRow label="Gap X" value="{spec.gapX}px" mono={false} />
    {/if}
    {#if spec?.gapY}
      <DescribeRow label="Gap Y" value="{spec.gapY}px" mono={false} />
    {/if}
  </DescribeSection>

  <!-- Components -->
  <DescribeSection title="Components ({componentCount})">
    {#each rows as row, rowIdx}
      {#each row.items ?? [] as item}
        <button
          class="flex items-baseline justify-between gap-x-4 min-h-[20px] w-full text-left hover:bg-surface-subtle rounded px-1 -mx-1 transition-colors"
          on:click={() => {
            if (item.component)
              dispatch("view-component", { componentName: item.component });
          }}
        >
          <span class="text-xs font-mono text-primary-600"
            >{item.component}</span
          >
          <span class="text-[10px] text-fg-muted">
            row {rowIdx + 1}{#if item.width}, w:{item.width}{/if}
          </span>
        </button>
      {/each}
    {/each}
  </DescribeSection>

  <!-- Variables -->
  <DescribeSection title="Variables">
    {#if variables.length > 0}
      {#each variables as v}
        <DescribeRow
          label={v.name ?? ""}
          value={String(v.defaultValue ?? "")}
        />
      {/each}
    {:else}
      <span class="text-xs text-fg-muted">None defined</span>
    {/if}
  </DescribeSection>

  <!-- Pinned Filters -->
  {#if spec?.pinnedFilters?.length}
    <DescribeSection title="Pinned Filters">
      <span class="text-xs font-mono text-fg-primary">
        {spec.pinnedFilters.join(", ")}
      </span>
    </DescribeSection>
  {/if}

  <!-- Security -->
  <DescribeSection title="Security Policy">
    {#if spec?.securityRules?.length}
      {#each spec.securityRules as rule, i}
        <div
          class="flex flex-col gap-y-1 {i > 0
            ? 'mt-1 pt-1 border-t border-border'
            : ''}"
        >
          {#if rule.access}
            <div class="flex flex-col gap-y-0.5">
              <span class="text-[11px] text-fg-secondary font-medium"
                >Access</span
              >
              <DescribeRow
                label={rule.access.allow ? "Allow" : "Deny"}
                value={rule.access.conditionExpression || "all"}
              />
              {#if rule.access.exclusive}
                <span class="text-[11px] text-fg-muted pl-2">exclusive</span>
              {/if}
            </div>
          {/if}
          {#if rule.rowFilter}
            <div class="flex flex-col gap-y-0.5">
              <span class="text-[11px] text-fg-secondary font-medium"
                >Row filter</span
              >
              {#if rule.rowFilter.sql}
                <span class="text-[11px] text-fg-muted font-mono pl-2"
                  >{rule.rowFilter.sql}</span
                >
              {/if}
              {#if rule.rowFilter.conditionExpression}
                <DescribeRow
                  label="Condition"
                  value={rule.rowFilter.conditionExpression}
                />
              {/if}
            </div>
          {/if}
          {#if rule.fieldAccess}
            <div class="flex flex-col gap-y-0.5">
              <span class="text-[11px] text-fg-secondary font-medium">
                Field {rule.fieldAccess.allow ? "include" : "exclude"}
              </span>
              {#if rule.fieldAccess.allFields}
                <span class="text-[11px] text-fg-muted pl-2">all fields</span>
              {:else if rule.fieldAccess.fields?.length}
                <span class="text-[11px] text-fg-muted font-mono pl-2">
                  {rule.fieldAccess.fields.join(", ")}
                </span>
              {/if}
              {#if rule.fieldAccess.conditionExpression}
                <DescribeRow
                  label="Condition"
                  value={rule.fieldAccess.conditionExpression}
                />
              {/if}
              {#if rule.fieldAccess.exclusive}
                <span class="text-[11px] text-fg-muted pl-2">exclusive</span>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    {:else}
      <span class="text-xs text-fg-muted">None defined</span>
    {/if}
  </DescribeSection>
</div>
