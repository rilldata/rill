<script lang="ts">
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import {
    ApplicationStore,
    duplicateSourceName,
    runtimeStore,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { config } from "@rilldata/web-local/lib/application-state-stores/application-store.js";
  import {
    importOverlayVisible,
    overlay,
    quickStartDashboardOverlay,
  } from "@rilldata/web-local/lib/application-state-stores/layout-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import DuplicateSource from "@rilldata/web-local/lib/components/navigation/sources/DuplicateSource.svelte";
  import NotificationCenter from "@rilldata/web-local/lib/components/notifications/NotificationCenter.svelte";
  import ExportingDataset from "@rilldata/web-local/lib/components/overlay/ExportingDataset.svelte";
  import FileDrop from "@rilldata/web-local/lib/components/overlay/FileDrop.svelte";
  import ImportingTable from "@rilldata/web-local/lib/components/overlay/ImportingTable.svelte";
  import PreparingImport from "@rilldata/web-local/lib/components/overlay/PreparingImport.svelte";
  import QuickStartDashboard from "@rilldata/web-local/lib/components/overlay/QuickStartDashboard.svelte";
  import ConfigProvider from "@rilldata/web-local/lib/config/ConfigProvider.svelte";
  import { initMetrics } from "@rilldata/web-local/lib/metrics/initMetrics";
  import {
    createQueryClient,
    queryClient,
  } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { getContext, onMount } from "svelte";
  import BlockingOverlayContainer from "../overlay/BlockingOverlayContainer.svelte";
  import BasicLayout from "./BasicLayout.svelte";
  createQueryClient();

  onMount(async () => {
    const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");

    runtimeStore.set({
      instanceId: instanceResp.instanceId,
      repoId: instanceResp.repoId,
    });

    return initMetrics();
  });

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  // FROM OLD INDEX.SVELTE

  let showDropOverlay = false;

  const app = getContext("rill:app:store") as ApplicationStore;
  $: debounceRunstate($app?.status || "disconnected");

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  // get any importing tables
  $: derivedImportedTable = $derivedTableStore?.entities?.find(
    (table) => table.status === EntityStatus.Importing
  );
  $: persistentImportedTable = $persistentTableStore?.entities?.find(
    (table) => table.id === derivedImportedTable?.id
  );
  // get any exporting datasets.
  $: derivedExportedModel = $derivedModelStore?.entities?.find(
    (model) => model.status === EntityStatus.Exporting
  );
  $: persistentExportedModel = $persistentModelStore?.entities?.find(
    (model) => model.id === derivedExportedModel?.id
  );

  function isEventWithFiles(event: DragEvent) {
    let types = event.dataTransfer.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<QueryClientProvider client={queryClient}>
  <ConfigProvider {config}>
    <div class="body">
      {#if derivedExportedModel && persistentExportedModel}
        <ExportingDataset tableName={persistentExportedModel.name} />
      {:else if derivedImportedTable && persistentImportedTable}
        <ImportingTable
          importName={persistentImportedTable.path}
          tableName={persistentImportedTable.name}
        />
      {:else if $importOverlayVisible}
        <PreparingImport />
      {:else if $quickStartDashboardOverlay?.show}
        <QuickStartDashboard
          sourceName={$quickStartDashboardOverlay.sourceName}
          timeDimension={$quickStartDashboardOverlay.timeDimension}
        />
      {:else if showDropOverlay}
        <FileDrop bind:showDropOverlay />
      {:else if $overlay !== null}
        <BlockingOverlayContainer
          bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
        >
          <div slot="title">
            <span class="font-bold">{$overlay?.title}</span>
          </div>
        </BlockingOverlayContainer>
      {/if}

      {#if $duplicateSourceName !== null}
        <DuplicateSource />
      {/if}

      <div
        class="index-body absolute w-screen h-screen"
        on:dragenter|preventDefault|stopPropagation
        on:dragleave|preventDefault|stopPropagation
        on:dragover|preventDefault|stopPropagation={(e) => {
          if (isEventWithFiles(e)) showDropOverlay = true;
        }}
        on:drag|preventDefault|stopPropagation
        on:drop|preventDefault|stopPropagation
      >
        <BasicLayout>
          <slot />
        </BasicLayout>
      </div>
    </div>
  </ConfigProvider>
</QueryClientProvider>

<NotificationCenter />
