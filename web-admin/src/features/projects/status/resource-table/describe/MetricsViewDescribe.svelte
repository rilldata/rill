<script lang="ts">
  import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";
  import SecurityRulesSection from "./SecurityRulesSection.svelte";
  import { formatTimeGrain, formatDayOfWeek, formatMonthOfYear } from "./utils";

  export let metricsView: V1MetricsView;

  $: spec = metricsView?.spec;
  $: state = metricsView?.state;
  $: dimensions = spec?.dimensions ?? [];
  $: measures = spec?.measures ?? [];
</script>

<div class="flex flex-col gap-y-3">
  <!-- Data Source -->
  {#if spec?.connector || spec?.table || spec?.model}
    <DescribeSection title="Data Source">
      <DescribeRow label="Connector" value={spec?.connector} />
      {#if spec?.model}
        <DescribeRow label="Model" value={spec.model} />
      {:else if spec?.table}
        <DescribeRow label="Table" value={spec.table} />
      {/if}
      <DescribeRow label="Database" value={spec?.database} />
      <DescribeRow label="Schema" value={spec?.databaseSchema} />
      <DescribeRow label="Parent" value={spec?.parent} />
      {#if state?.dataRefreshedOn}
        <DescribeRow
          label="Data refreshed on"
          value={new Date(state.dataRefreshedOn).toLocaleString()}
          mono={false}
        />
      {/if}
    </DescribeSection>
  {/if}

  <!-- Time -->
  <DescribeSection title="Time">
    {#if spec?.timeDimension}
      <DescribeRow label="Time dimension" value={spec.timeDimension} />
    {:else}
      <DescribeRow
        label="Time dimension"
        value="Inferred from time series"
        mono={false}
      />
    {/if}
    <DescribeRow
      label="Smallest time grain"
      value={formatTimeGrain(spec?.smallestTimeGrain)}
    />
    <DescribeRow
      label="Watermark expression"
      value={spec?.watermarkExpression}
    />
    <DescribeRow
      label="First day of week"
      value={formatDayOfWeek(spec?.firstDayOfWeek)}
      mono={false}
    />
    <DescribeRow
      label="First month of year"
      value={formatMonthOfYear(spec?.firstMonthOfYear)}
      mono={false}
    />
  </DescribeSection>

  <!-- Dimensions -->
  <DescribeSection title="Dimensions ({dimensions.length})">
    {#each dimensions as dim (dim.name)}
      <div class="flex flex-col gap-y-0.5">
        <div class="flex items-baseline justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs font-mono text-fg-primary">{dim.name}</span>
          {#if dim.type && dim.type !== "DIMENSION_TYPE_UNSPECIFIED"}
            {@const label = dim.type.replace("DIMENSION_TYPE_", "")}
            {#if label !== "SIMPLE"}
              <span class="text-[10px] text-fg-muted">{label}</span>
            {/if}
          {/if}
        </div>
        {#if dim.description}
          <span class="text-[11px] text-fg-secondary pl-2"
            >{dim.description}</span
          >
        {/if}
        {#if dim.expression}
          <span class="text-[11px] text-fg-muted font-mono pl-2"
            >{dim.expression}</span
          >
        {:else if dim.column}
          <span class="text-[11px] text-fg-muted font-mono pl-2"
            >column: {dim.column}</span
          >
        {/if}
        {#if dim.unnest}
          <span class="text-[11px] text-fg-muted pl-2">unnest</span>
        {/if}
      </div>
    {/each}
  </DescribeSection>

  <!-- Measures -->
  <DescribeSection title="Measures ({measures.length})">
    {#each measures as m (m.name)}
      <div class="flex flex-col gap-y-0.5">
        <div class="flex items-baseline justify-between gap-x-4 min-h-[20px]">
          <span class="text-xs font-mono text-fg-primary">{m.name}</span>
          {#if m.type && m.type !== "MEASURE_TYPE_UNSPECIFIED"}
            {@const label = m.type.replace("MEASURE_TYPE_", "")}
            {#if label !== "SIMPLE"}
              <span class="text-[10px] text-fg-muted">{label}</span>
            {/if}
          {/if}
        </div>
        {#if m.description}
          <span class="text-[11px] text-fg-secondary pl-2">{m.description}</span
          >
        {/if}
        {#if m.expression}
          <span class="text-[11px] text-fg-muted font-mono pl-2"
            >{m.expression}</span
          >
        {/if}
        {#if m.formatPreset || m.formatD3}
          <span class="text-[11px] text-fg-muted pl-2"
            >format: {m.formatPreset || m.formatD3}</span
          >
        {/if}
        {#if m.window}
          <span class="text-[11px] text-fg-muted pl-2">windowed</span>
        {/if}
        {#if m.validPercentOfTotal}
          <span class="text-[11px] text-fg-muted pl-2">valid % of total</span>
        {/if}
      </div>
    {/each}
  </DescribeSection>

  <!-- Caching -->
  {#if spec?.cacheEnabled !== undefined}
    <DescribeSection title="Cache">
      <DescribeRow label="Enabled" value={spec.cacheEnabled ? "Yes" : "No"} />
      <DescribeRow label="Cache key SQL" value={spec.cacheKeySql} />
      <DescribeRow
        label="Cache TTL (seconds)"
        value={spec.cacheKeyTtlSeconds}
        mono={false}
      />
    </DescribeSection>
  {/if}

  <!-- AI Instructions -->
  {#if spec?.aiInstructions}
    <DescribeSection title="AI Instructions">
      <span class="text-xs text-fg-primary">
        {spec.aiInstructions}
      </span>
    </DescribeSection>
  {/if}

  <!-- Security -->
  <SecurityRulesSection rules={spec?.securityRules} />

  <!-- Annotations -->
  {#if spec?.annotations?.length}
    <DescribeSection title="Annotations">
      {#each spec.annotations as annotation, i (annotation.name ?? i)}
        <DescribeRow label="Annotation {i + 1}" value={annotation.name} />
        <DescribeRow label="  Table" value={annotation.table} />
        <DescribeRow label="  Model" value={annotation.model} />
        {#if annotation.measures?.length}
          <DescribeRow
            label="  Measures"
            value={annotation.measures.join(", ")}
          />
        {/if}
      {/each}
    </DescribeSection>
  {/if}
</div>
