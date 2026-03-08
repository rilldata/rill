<script lang="ts">
  import { goto } from "$app/navigation";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    runtimeServiceImportDbtMetrics,
    createRuntimeServiceImportDbtMetricsMutation,
  } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
  import { dbtImportModal } from "./dbt-import-modal-store";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  const client = useRuntimeClient();

  // Metric listing state
  let metrics: Array<{
    name: string;
    label: string;
    description: string;
    type: string;
  }> = [];
  let loading = true;
  let listError = "";

  // Warehouse connector state (auto-detected from manifest adapter type)
  let adapterType = "";
  let matchingConnectors: string[] = [];
  let selectedWarehouse = "";

  // Selection state
  let selected = new Set<string>();

  // Import mutation
  const importMutation = createRuntimeServiceImportDbtMetricsMutation(client);

  $: importError = $importMutation.isError
    ? String(
        ($importMutation.error as Record<string, unknown>)?.message ??
          $importMutation.error,
      )
    : "";

  // Fetch metrics when modal opens
  $: if ($dbtImportModal.open && $dbtImportModal.connectorName) {
    fetchMetrics($dbtImportModal.connectorName);
  }

  async function fetchMetrics(connectorName: string) {
    loading = true;
    listError = "";
    metrics = [];
    selected = new Set();
    adapterType = "";
    matchingConnectors = [];
    selectedWarehouse = "";

    try {
      const resp = await runtimeServiceImportDbtMetrics(client, {
        connector: connectorName,
        listOnly: true,
      });

      adapterType = (resp.adapterType as string) ?? "";
      matchingConnectors = ((resp.matchingConnectors ?? []) as string[]).filter(
        Boolean,
      );

      // Auto-select if exactly one matching connector
      if (matchingConnectors.length === 1) {
        selectedWarehouse = matchingConnectors[0];
      }

      const available = (resp.availableMetrics ?? []) as Array<{
        name?: string;
        label?: string;
        description?: string;
        type?: string;
      }>;
      metrics = available.map((m) => ({
        name: m.name ?? "",
        label: m.label ?? "",
        description: m.description ?? "",
        type: m.type ?? "",
      }));
      // Start with none selected
      selected = new Set();
    } catch (e) {
      const err = e as Record<string, unknown>;
      listError = String(err?.message ?? e);
    } finally {
      loading = false;
    }
  }

  function toggleAll() {
    if (selected.size === metrics.length) {
      selected = new Set();
    } else {
      selected = new Set(metrics.map((m) => m.name));
    }
  }

  function toggleMetric(name: string) {
    const next = new Set(selected);
    if (next.has(name)) {
      next.delete(name);
    } else {
      next.add(name);
    }
    selected = next;
  }

  async function handleImport() {
    const refs = Array.from(selected);
    if (refs.length === 0 || !selectedWarehouse) return;

    await $importMutation.mutateAsync({
      connector: $dbtImportModal.connectorName,
      metricRefs: refs,
      warehouseConnector: selectedWarehouse,
    });

    const files = ($importMutation.data?.generatedFiles ?? []) as string[];
    const count = Math.floor(files.length / 2); // each metric generates model + metrics_view
    eventBus.emit("notification", {
      message: `Imported ${count} metric${count !== 1 ? "s" : ""}`,
    });

    dbtImportModal.close();

    // Navigate to first generated metrics view
    const firstMv = files.find((f) => f.startsWith("/metrics/"));
    if (firstMv) {
      await goto(`/files${firstMv}`);
    }
  }

  function close() {
    if (!$importMutation.isPending) {
      dbtImportModal.close();
    }
  }

  $: allSelected = metrics.length > 0 && selected.size === metrics.length;
  $: connectorOptions = matchingConnectors.map((name) => ({
    value: name,
    label: name,
  }));
  $: canImport =
    selected.size > 0 &&
    selectedWarehouse !== "" &&
    !loading &&
    !$importMutation.isPending;
</script>

<Dialog.Root
  open={$dbtImportModal.open}
  onOpenChange={(open) => {
    if (!open) close();
  }}
>
  <Dialog.Content class="max-w-lg">
    <Dialog.Title>Import dbt metrics</Dialog.Title>

    <Dialog.Description>
      Select metrics to import from <span class="font-mono"
        >{$dbtImportModal.connectorName}</span
      >. Each metric creates a model and metrics view.
    </Dialog.Description>

    <div class="mt-4 flex flex-col gap-3 max-h-80 min-h-[120px]">
      {#if loading}
        <div class="flex items-center justify-center py-8 gap-2">
          <LoadingSpinner size="16px" />
          <span class="text-sm text-fg-secondary">Loading metrics...</span>
        </div>
      {:else if listError}
        <div class="p-3 bg-red-50 border border-red-200 rounded-md text-sm">
          <p class="text-red-800 font-medium">Failed to load metrics</p>
          <p class="text-red-600 mt-1">{listError}</p>
        </div>
      {:else if metrics.length === 0}
        <div class="py-8 text-center text-sm text-fg-secondary">
          No metrics found in the dbt manifest.
        </div>
      {:else}
        <button
          type="button"
          class="flex items-center gap-2 border-b border-gray-200 pb-2 w-full text-left"
          on:click={toggleAll}
        >
          <div class="pointer-events-none shrink-0">
            <Checkbox checked={allSelected} />
          </div>
          <span class="text-sm">Select all ({metrics.length})</span>
        </button>

        <div class="overflow-y-auto flex flex-col gap-1">
          {#each metrics as metric (metric.name)}
            <button
              type="button"
              class="flex items-start gap-2 px-2 py-1.5 rounded hover:bg-gray-50 text-left w-full"
              on:click={() => toggleMetric(metric.name)}
            >
              <div class="pt-0.5 shrink-0 pointer-events-none">
                <Checkbox checked={selected.has(metric.name)} />
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-sm font-medium text-fg-primary truncate">
                  {metric.label || metric.name}
                </span>
                {#if metric.description}
                  <span class="text-xs text-fg-secondary line-clamp-2">
                    {metric.description}
                  </span>
                {/if}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    {#if !loading && !listError && metrics.length > 0}
      <div class="mt-3 border-t border-gray-200 pt-3">
        {#if matchingConnectors.length === 0}
          <div class="p-3 bg-red-50 border border-red-200 rounded-md text-sm">
            <p class="text-red-800 font-medium">No matching connector found</p>
            <p class="text-red-600 mt-1">
              {#if adapterType}
                No Rill connector found for warehouse type "{adapterType}".
                Please set up a {adapterType} connector first.
              {:else}
                Could not detect warehouse type from the dbt manifest.
              {/if}
            </p>
          </div>
        {:else if matchingConnectors.length === 1}
          <p class="text-sm text-fg-secondary">
            Warehouse connector: <span class="font-medium text-fg-primary"
              >{matchingConnectors[0]}</span
            >
            {#if adapterType}
              <span class="text-fg-disabled">({adapterType})</span>
            {/if}
          </p>
        {:else}
          <Select
            id="warehouse-connector"
            bind:value={selectedWarehouse}
            options={connectorOptions}
            label="Warehouse connector"
            placeholder="Select a connector"
            tooltip="The Rill connector for the warehouse where dbt models are materialized ({adapterType})"
            full
          />
        {/if}
      </div>
    {/if}

    {#if importError}
      <div
        class="mt-3 p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800"
      >
        {importError}
      </div>
    {/if}

    <Dialog.Footer>
      <Button
        type="secondary"
        onClick={close}
        disabled={$importMutation.isPending}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={handleImport}
        disabled={!canImport}
        loading={$importMutation.isPending}
        loadingCopy="Importing..."
      >
        Import {selected.size} metric{selected.size !== 1 ? "s" : ""}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
