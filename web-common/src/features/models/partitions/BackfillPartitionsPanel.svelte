<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { type ColumnDef, renderComponent } from "tanstack-table-8-svelte-5";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetModelPartitionsInfinite,
    type V1ModelPartition,
    type V1Resource,
  } from "../../../runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import DataCell from "./DataCell.svelte";
  import ErrorCell from "./ErrorCell.svelte";
  import {
    filterPartitionsInRange,
    getPartitionExecutionSettings,
    getPartitionFields,
    isPartitionPending,
    validatePartitionRange,
    type PartitionRangeError,
  } from "./partition-range";

  export let resource: V1Resource;
  export let onDone: () => void;
  export let onTriggered: () => void;

  const runtimeClient = useRuntimeClient();
  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  let field = "";
  let start = "";
  let end = "";
  let rangeError: PartitionRangeError | undefined = undefined;
  let matched: V1ModelPartition[] | null = null;
  let submitError = "";

  $: modelName = resource.meta?.name?.name as string;

  $: query = createRuntimeServiceGetModelPartitionsInfinite(
    runtimeClient,
    { model: modelName },
    { query: { enabled: !!modelName, refetchOnMount: true } },
  );

  // Fetch every page so the range covers all partitions.
  $: if ($query.hasNextPage && !$query.isFetchingNextPage && !$query.isError) {
    void $query.fetchNextPage();
  }

  $: allRows =
    $query.data?.pages.flatMap((p) => p.partitions as V1ModelPartition[]) ?? [];
  $: isLoadingAll = $query.isLoading || ($query.hasNextPage && !$query.isError);
  $: isReconciling =
    resource.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
  $: fieldOptions = getPartitionFields(allRows).map((f) => ({
    value: f,
    label: f,
  }));
  $: settings = getPartitionExecutionSettings(
    resource.model?.spec,
    resource.model?.state,
  );
  $: matchedKeys = new Set((matched ?? []).map((p) => p.key));
  $: otherPendingCount = matched
    ? allRows.filter((p) => isPartitionPending(p) && !matchedKeys.has(p.key))
        .length
    : 0;

  // Any edit to the range invalidates a previous preview.
  $: field, start, end, ((matched = null), (rangeError = undefined));

  function preview() {
    submitError = "";
    rangeError = validatePartitionRange(field, start.trim(), end.trim());
    if (rangeError) return;
    matched = filterPartitionsInRange(allRows, field, start.trim(), end.trim());
  }

  async function backfill() {
    if (!matched?.length) return;
    submitError = "";
    try {
      await $triggerMutation.mutateAsync({
        models: [
          {
            model: modelName,
            partitions: matched.map((p) => p.key as string),
          },
        ],
      });
      onTriggered();
    } catch (e) {
      submitError = m.status_backfill_failed({
        message: e instanceof Error ? e.message : String(e),
      });
    }
  }

  const columns: ColumnDef<V1ModelPartition, any>[] = [
    {
      id: "data",
      header: "Data",
      cell: ({ row }) => renderComponent(DataCell, { data: row.original.data }),
      enableSorting: false,
    },
    {
      accessorKey: "executedOn",
      header: "Executed on",
      cell: ({ row }) =>
        row.original.skipped
          ? "Skipped"
          : row.original.executedOn
            ? new Date(row.original.executedOn).toLocaleString()
            : "-",
    },
    {
      accessorKey: "error",
      header: "Error",
      cell: ({ row }) =>
        renderComponent(ErrorCell, { error: row.original.error }),
      enableSorting: false,
    },
  ];
</script>

<div class="flex flex-col gap-y-4 min-h-0 flex-1">
  <div class="flex flex-row items-end gap-x-3">
    <div class="w-56">
      <Select
        id="backfill-field"
        label={m.status_backfill_field_label()}
        placeholder={m.status_backfill_field_placeholder()}
        options={fieldOptions}
        optionsLoading={isLoadingAll}
        bind:value={field}
        size="md"
        full
      />
    </div>
    <div class="w-56">
      <Input
        id="backfill-start"
        label={m.status_backfill_start_label()}
        placeholder="2026-03-01"
        bind:value={start}
        size="md"
      />
    </div>
    <div class="w-56">
      <Input
        id="backfill-end"
        label={m.status_backfill_end_label()}
        placeholder="2026-03-07"
        bind:value={end}
        size="md"
      />
    </div>
    <Button type="secondary" onClick={preview} disabled={isLoadingAll}>
      {m.status_backfill_preview()}
    </Button>
  </div>

  {#if $query.error}
    <span class="text-xs text-red-600">
      {m.common_error_message({ message: $query.error.message })}
    </span>
  {/if}

  {#if rangeError}
    <span class="text-xs text-red-600">
      {rangeError === "missing"
        ? m.status_backfill_error_missing()
        : m.status_backfill_error_start_after_end()}
    </span>
  {/if}

  <div class="settings">
    <div class="flex flex-row items-baseline justify-between">
      <span class="font-semibold text-fg-primary">
        {m.status_backfill_settings_title()}
      </span>
      <span class="text-fg-muted">{m.status_backfill_settings_hint()}</span>
    </div>
    <div class="settings-row">
      <span class="settings-label">{m.status_backfill_concurrency_label()}</span
      >
      <span class="text-fg-primary">
        {#if settings.concurrency !== undefined}
          {m.status_backfill_concurrency_value({
            count: settings.concurrency,
          })}
        {/if}
        <span class="text-fg-muted">
          {settings.concurrencySource === "observed"
            ? m.status_backfill_concurrency_observed()
            : settings.concurrencySource === "configured"
              ? m.status_backfill_concurrency_configured()
              : m.status_backfill_concurrency_unknown()}
        </span>
      </span>
    </div>
    <div class="settings-row">
      <span class="settings-label">{m.status_backfill_retry_label()}</span>
      <span class="text-fg-primary">
        {m.status_backfill_retry_attempts({ count: settings.retryAttempts })},
        {m.status_backfill_retry_delay({
          seconds: settings.retryDelaySeconds,
        })},
        {settings.retryExponentialBackoff
          ? m.status_backfill_retry_backoff()
          : m.status_backfill_retry_fixed()}
        <span class="text-fg-muted">
          {settings.retryIsDefault
            ? m.status_backfill_retry_default()
            : m.status_backfill_retry_configured()}
        </span>
      </span>
    </div>
    <div class="settings-row">
      <span class="settings-label"></span>
      <span class="text-fg-secondary break-all">
        {m.status_backfill_retry_matches({
          patterns: settings.retryIfErrorMatches.join(", "),
        })}
      </span>
    </div>
  </div>

  {#if matched}
    <div class="flex flex-col gap-y-1 text-sm">
      <span class="font-semibold text-fg-primary">
        {matched.length
          ? m.status_backfill_matched({ count: matched.length })
          : m.status_backfill_no_match()}
      </span>
      {#if matched.length && otherPendingCount > 0}
        <span class="text-xs text-yellow-600">
          {m.status_backfill_pending_note({ count: otherPendingCount })}
        </span>
      {/if}
    </div>
    {#if matched.length}
      <div class="flex-1 min-h-0 overflow-auto">
        <BasicTable
          data={matched}
          {columns}
          columnLayout="minmax(0, 3fr) minmax(0, 1fr) minmax(0, 2fr)"
        />
      </div>
    {/if}
  {/if}

  {#if submitError}
    <span class="text-xs text-red-600">{submitError}</span>
  {/if}

  <div class="flex flex-row justify-end gap-x-2">
    <Button type="tertiary" onClick={onDone}>{m.status_backfill_back()}</Button>
    <Button
      type="primary"
      onClick={backfill}
      disabled={!matched?.length || isReconciling || $triggerMutation.isPending}
      loading={$triggerMutation.isPending}
    >
      {m.status_backfill_confirm({ count: matched?.length ?? 0 })}
    </Button>
  </div>
</div>

<style lang="postcss">
  .settings {
    @apply flex flex-col gap-y-1.5 rounded-sm border p-3 text-xs bg-surface-subtle;
  }
  .settings-row {
    @apply flex flex-row gap-x-3;
  }
  .settings-label {
    @apply w-24 shrink-0 text-fg-secondary;
  }
</style>
