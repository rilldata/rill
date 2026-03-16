<script lang="ts">
  import type { V1Explore } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";
  import SecurityRulesSection from "./SecurityRulesSection.svelte";

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
    <DescribeRow label="Lock time zone" value={String(!!spec?.lockTimeZone)} />
    <DescribeRow
      label="Allow custom time range"
      value={String(!!spec?.allowCustomTimeRange)}
    />
    <DescribeRow
      label="Hide pivot in embeds"
      value={String(!!spec?.embedsHidePivot)}
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

  <!-- Security -->
  <SecurityRulesSection rules={spec?.securityRules} />
</div>
