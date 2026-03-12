<script lang="ts">
  import type { V1Explore } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";

  export let explore: V1Explore;

  $: spec = explore?.spec;
  $: state = explore?.state;
  $: dimensions = spec?.dimensions ?? [];
  $: measures = spec?.measures ?? [];
</script>

<div class="flex flex-col gap-y-3">
  <!-- General -->
  <DescribeSection title="General">
    <DescribeRow label="Metrics view" value={spec?.metricsView} />
    {#if spec?.theme}
      <DescribeRow label="Theme" value={spec.theme} />
    {/if}
    {#if spec?.banner}
      <DescribeRow label="Banner" value={spec.banner} mono={false} />
    {/if}
    <DescribeRow
      label="Defined in metrics view"
      value={String(!!spec?.definedInMetricsView)}
    />
    {#if state?.dataRefreshedOn}
      <DescribeRow
        label="Data refreshed on"
        value={new Date(state.dataRefreshedOn).toLocaleString()}
        mono={false}
      />
    {/if}
  </DescribeSection>

  <!-- Dimensions -->
  <DescribeSection title="Dimensions">
    {#if dimensions.length > 0}
      <span class="text-xs font-mono text-fg-primary">
        {dimensions.join(", ")}
      </span>
    {:else}
      <span class="text-xs text-fg-muted">All (from metrics view)</span>
    {/if}
  </DescribeSection>

  <!-- Measures -->
  <DescribeSection title="Measures">
    {#if measures.length > 0}
      <span class="text-xs font-mono text-fg-primary">
        {measures.join(", ")}
      </span>
    {:else}
      <span class="text-xs text-fg-muted">All (from metrics view)</span>
    {/if}
  </DescribeSection>

  <!-- Options -->
  <DescribeSection title="Options">
    <DescribeRow
      label="Lock time zone"
      value={String(!!spec?.lockTimeZone)}
    />
    <DescribeRow
      label="Allow custom time range"
      value={String(!!spec?.allowCustomTimeRange)}
    />
    <DescribeRow
      label="Hide pivot in embeds"
      value={String(!!spec?.embedsHidePivot)}
    />
  </DescribeSection>

  <!-- Security -->
  <DescribeSection title="Security Policy">
    {#if spec?.securityRules?.length}
      {#each spec.securityRules as rule, i}
        <div class="flex flex-col gap-y-1 {i > 0 ? 'mt-1 pt-1 border-t border-border' : ''}">
          {#if rule.access}
            <div class="flex flex-col gap-y-0.5">
              <span class="text-[11px] text-fg-secondary font-medium">Access</span>
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
              <span class="text-[11px] text-fg-secondary font-medium">Row filter</span>
              {#if rule.rowFilter.sql}
                <span class="text-[11px] text-fg-muted font-mono pl-2">{rule.rowFilter.sql}</span>
              {/if}
              {#if rule.rowFilter.conditionExpression}
                <DescribeRow label="Condition" value={rule.rowFilter.conditionExpression} />
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
                <DescribeRow label="Condition" value={rule.fieldAccess.conditionExpression} />
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
