<script lang="ts">
  import type { V1Source, V1Model } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";
  import { formatSchedule, formatBytes, formatChangeMode } from "./utils";

  export let source: V1Source | undefined = undefined;
  export let model: V1Model | undefined = undefined;

  $: isSource = !!source;
  $: spec = isSource ? source?.spec : model?.spec;

  // Source-specific
  $: sourceSpec = source?.spec;
  $: sourceState = source?.state;

  // Model-specific
  $: modelSpec = model?.spec;
  $: modelState = model?.state;

  // Connection info
  $: inputConnector = isSource
    ? sourceSpec?.sourceConnector
    : modelSpec?.inputConnector;
  $: outputConnector = isSource
    ? sourceSpec?.sinkConnector
    : modelSpec?.outputConnector;

  // Schedule
  $: schedule = spec?.refreshSchedule;

  // Refresh
  $: refreshedOn = isSource
    ? sourceState?.refreshedOn
    : modelState?.refreshedOn;

  // SQL (from input properties for models, or source properties)
  $: sql = isSource
    ? (sourceSpec?.properties?.sql as string | undefined)
    : (modelSpec?.inputProperties?.sql as string | undefined);

  // Materialize (from outputProperties.materialize)
  $: materialize = !isSource
    ? (modelSpec?.outputProperties?.materialize as boolean | undefined)
    : undefined;

  // Incremental & partitioned flags
  $: incremental = !isSource ? !!modelSpec?.incremental : undefined;
  $: partitioned = !isSource ? !!modelSpec?.partitionsResolver : undefined;

  // Source properties (excluding sql)
  $: sourceProperties = sourceSpec?.properties ?? {};
  $: sourcePropertyKeys = Object.keys(sourceProperties)
    .filter((k) => k !== "sql")
    .sort();
</script>

<div class="flex flex-col gap-y-3">
  <!-- General -->
  <DescribeSection title="General">
    {#if refreshedOn}
      <DescribeRow
        label="Last refreshed"
        value={new Date(refreshedOn).toLocaleString()}
        mono={false}
      />
    {/if}
    {#if !isSource && modelSpec?.changeMode}
      <DescribeRow
        label="Change mode"
        value={formatChangeMode(modelSpec.changeMode)}
      />
    {/if}
    <DescribeRow
      label="Input / Output"
      value="{inputConnector} / {outputConnector}"
    />
    {#if schedule && !schedule.disable}
      <DescribeRow label="Refresh" value={formatSchedule(schedule)} />
      {#if schedule.timeZone}
        <DescribeRow label="Time zone" value={schedule.timeZone} />
      {/if}
    {/if}
    {#if materialize !== undefined}
      <DescribeRow label="Materialize" value={String(materialize)} />
    {/if}
    {#if incremental !== undefined}
      <DescribeRow label="Incremental" value={String(incremental)} />
    {/if}
    {#if partitioned !== undefined}
      <DescribeRow label="Partitioned" value={String(partitioned)} />
    {/if}
  </DescribeSection>

  <!-- SQL -->
  {#if sql}
    <DescribeSection title="SQL">
      <pre
        class="text-xs font-mono whitespace-pre-wrap text-fg-primary overflow-auto max-h-[30vh] leading-relaxed">{sql}</pre>
    </DescribeSection>
  {/if}

  <!-- Source properties -->
  {#if isSource && sourcePropertyKeys.length > 0}
    <DescribeSection title="Properties">
      {#each sourcePropertyKeys as key}
        <DescribeRow label={key} value={String(sourceProperties[key])} />
      {/each}
    </DescribeSection>
  {/if}

  <!-- Model input/stage/output properties -->
  {#if !isSource}
    {#if modelSpec?.inputProperties && Object.keys(modelSpec.inputProperties).filter((k) => k !== "sql").length > 0}
      <DescribeSection title="Input Properties">
        {#each Object.entries(modelSpec.inputProperties).filter(([k]) => k !== "sql") as [key, val]}
          <DescribeRow label={key} value={String(val)} />
        {/each}
      </DescribeSection>
    {/if}

    {#if modelSpec?.stageConnector}
      <DescribeSection title="Stage">
        <DescribeRow label="Connector" value={modelSpec.stageConnector} />
        {#if modelSpec?.stageProperties && Object.keys(modelSpec.stageProperties).length > 0}
          {#each Object.entries(modelSpec.stageProperties) as [key, val]}
            <DescribeRow label={key} value={String(val)} />
          {/each}
        {/if}
      </DescribeSection>
    {/if}
  {/if}
  <!-- Runtime Info -->
  <DescribeSection title="Runtime">
    {#if isSource}
      <DescribeRow label="Connector" value={sourceState?.connector} />
    {:else if modelState}
      {#if modelState.rowsTotal}
        <DescribeRow
          label="Rows"
          value={Number(modelState.rowsTotal).toLocaleString()}
          mono={false}
        />
      {/if}
      {#if modelState.bytesTotal}
        <DescribeRow
          label="Size"
          value={formatBytes(modelState.bytesTotal)}
          mono={false}
        />
      {/if}
      {#if modelState.partitionsModelId}
        <DescribeRow
          label="Partitions model ID"
          value={modelState.partitionsModelId}
        />
      {/if}
      {#if modelState.partitionsHaveErrors}
        <DescribeRow label="Partitions have errors" value="Yes" />
      {/if}
      {#if modelState.totalExecutionDurationMs}
        <DescribeRow
          label="Total execution duration"
          value="{Number(
            modelState.totalExecutionDurationMs,
          ).toLocaleString()} ms"
          mono={false}
        />
      {/if}
      {#if modelState.latestExecutionDurationMs}
        <DescribeRow
          label="Latest execution duration"
          value="{Number(
            modelState.latestExecutionDurationMs,
          ).toLocaleString()} ms"
          mono={false}
        />
      {/if}
    {/if}
  </DescribeSection>

  <!-- Retry (model only) -->
  {#if !isSource && modelSpec?.retryAttempts}
    <DescribeSection title="Retry">
      <DescribeRow
        label="Attempts"
        value={modelSpec.retryAttempts}
        mono={false}
      />
      <DescribeRow
        label="Delay (seconds)"
        value={modelSpec.retryDelaySeconds}
        mono={false}
      />
      {#if modelSpec.retryIfErrorMatches?.length}
        <div class="flex flex-col gap-y-0.5">
          <span class="text-xs text-fg-secondary">Error match patterns</span>
          {#each modelSpec.retryIfErrorMatches as pattern}
            <span class="text-xs font-mono text-fg-primary text-right"
              >{pattern}</span
            >
          {/each}
        </div>
      {/if}
    </DescribeSection>
  {/if}

  <!-- Tests (model only) -->
  {#if !isSource && modelSpec?.tests?.length}
    <DescribeSection title="Tests ({modelSpec.tests.length})">
      {#each modelSpec.tests as test}
        <DescribeRow label={test.name ?? "test"} value={test.resolver} />
      {/each}
    </DescribeSection>
  {/if}
</div>
