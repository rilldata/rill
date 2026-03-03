<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { connectorIconMapping } from "../../connectors/connector-icon-mapping";
  import { createRuntimeServiceAnalyzeConnectors } from "../../../runtime-client";
  import { ResourceKind } from "../../entity-management/resource-selectors";
  import { createResourceAndNavigate } from "../../file-explorer/new-files";
  import AddDataForm from "./AddDataForm.svelte";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import {
    modelConnectors,
    getBackendConnectorName,
    getConnectorSchema,
    getFormWidth,
    hasExplorerStep as hasExplorerStepSchema,
    isMultiStepConnector as isMultiStepConnectorSchema,
    type ConnectorInfo,
  } from "./connector-schemas";
  import { ICONS } from "./icons";
  import {
    connectorStepStore,
    resetConnectorStep,
    setStep,
    setConnectorInstanceName as setStoreConnectorInstanceName,
  } from "./connectorStepStore";
  import { File as FileIcon } from "lucide-svelte";

  let step = 0;
  let selectedConnector: null | V1ConnectorDriver = null;
  let selectedSchemaName: string | null = null;
  let pendingConnectorName: string | null = null;
  let connectorInstanceName: string | null = null;
  let isSubmittingForm = false;

  // Track the previous connector step to detect back-navigation transitions
  let prevConnectorStep: string | null = null;

  $: ({ instanceId } = $runtime);

  // Fetch existing connector instances (SQL stores, warehouses, object stores)
  $: connectorInstances = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      select: (data) => {
        if (!data?.connectors) return { connectors: [] };
        const filtered = data.connectors
          .filter(
            (c) =>
              !c?.errorMessage &&
              !c?.driver?.implementsOlap &&
              (c?.driver?.implementsSqlStore ||
                c?.driver?.implementsWarehouse ||
                c?.driver?.implementsObjectStore),
          )
          .sort((a, b) => (a?.name as string).localeCompare(b?.name as string));
        return { connectors: filtered };
      },
    },
  });
  $: existingInstances = $connectorInstances.data?.connectors ?? [];

  // Deduplicate: get driver names covered by existing instances
  $: existingDriverNames = new Set(
    existingInstances.map((c) => c.driver?.name).filter(Boolean),
  );

  // Filter model connectors to exclude types already covered by existing instances
  // Always include "public" and "local_file" since they aren't typical connector instances
  $: uncoveredModelConnectors = modelConnectors.filter((c) => {
    const backendName = getBackendConnectorName(c.name);
    if (c.name === "public" || c.name === "local_file") return true;
    return !existingDriverNames.has(backendName);
  });

  // Intercept back-navigation from the import step.
  // When handleBack() in AddDataFormManager transitions from source/explorer
  // back to connector step, we redirect to the tile selection instead of
  // showing the connector credentials UI.
  $: {
    const currentConnectorStep = $connectorStepStore.step;
    if (
      step === 2 &&
      prevConnectorStep !== null &&
      (prevConnectorStep === "source" || prevConnectorStep === "explorer") &&
      currentConnectorStep === "connector"
    ) {
      goBackToTileSelection();
    }
    prevConnectorStep = currentConnectorStep;
  }

  // Get the form width class for the selected connector
  $: selectedSchema = selectedSchemaName
    ? getConnectorSchema(selectedSchemaName)
    : null;
  $: formWidthClass = getFormWidth(selectedSchema);

  /**
   * Convert a ConnectorInfo (from schema) to a V1ConnectorDriver-compatible object.
   */
  function toConnectorDriver(info: ConnectorInfo): V1ConnectorDriver {
    const schema = getConnectorSchema(info.name);
    const category = schema?.["x-category"];
    const backendName = getBackendConnectorName(info.name);

    return {
      name: backendName,
      displayName: info.displayName,
      implementsObjectStore: category === "objectStore",
      implementsOlap: category === "olap",
      implementsSqlStore: category === "sqlStore",
      implementsWarehouse: category === "warehouse",
      implementsFileStore: category === "fileStore",
    };
  }

  onMount(() => {
    function listen(e: PopStateEvent) {
      // Only handle events from the AddModelModal
      if (e.state?.modal !== "model") return;

      const stateStep = e.state?.step ?? 0;
      connectorInstanceName = e.state?.connectorInstanceName ?? null;

      // Handle both full connector object and connector name string
      if (e.state?.selectedConnector) {
        selectedConnector = e.state.selectedConnector;
        selectedSchemaName = e.state?.schemaName ?? null;
        pendingConnectorName = null;
        step = stateStep;
      } else if (e.state?.connector) {
        pendingConnectorName = e.state.connector;
        selectedSchemaName = e.state.connector;
        step = 2;
        // Try to resolve immediately if connectors are already loaded
        const found = modelConnectors.find((c) => c.name === e.state.connector);
        if (found) {
          selectedConnector = toConnectorDriver(found);
          pendingConnectorName = null;
        }
      } else {
        selectedConnector = null;
        selectedSchemaName = null;
        pendingConnectorName = null;
        step = stateStep;
      }
    }
    window.addEventListener("popstate", listen);

    return () => {
      window.removeEventListener("popstate", listen);
    };
  });

  // Handle pending connector name when connectors finish loading
  $: if (pendingConnectorName && modelConnectors.length > 0) {
    const found = modelConnectors.find((c) => c.name === pendingConnectorName);
    if (found) {
      selectedConnector = toConnectorDriver(found);
      selectedSchemaName = pendingConnectorName;
      pendingConnectorName = null;
      step = 2;
    }
  }

  function goToConnectorForm(connectorInfo: ConnectorInfo) {
    resetConnectorStep();

    // For multi-step connectors, skip directly to the import/source step.
    // For SQL connectors with explorer, skip to explorer step.
    const schema = getConnectorSchema(connectorInfo.name);
    if (isMultiStepConnectorSchema(schema)) {
      setStep("source");
    } else if (hasExplorerStepSchema(schema)) {
      setStep("explorer");
    }

    const state = {
      modal: "model" as const,
      step: 2,
      selectedConnector: toConnectorDriver(connectorInfo),
      schemaName: connectorInfo.name,
      connectorInstanceName: null,
    };
    window.history.pushState(state, "", "");
    window.dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  function goToExistingConnector(instance: {
    name?: string;
    driver?: V1ConnectorDriver;
  }) {
    const driverName = instance.driver?.name ?? "";
    const schema = getConnectorSchema(driverName);
    const hasExplorer = hasExplorerStepSchema(schema);
    const targetStep = hasExplorer ? "explorer" : "source";

    resetConnectorStep();
    setStep(targetStep);
    setStoreConnectorInstanceName(instance.name ?? null);

    const state = {
      modal: "model" as const,
      step: 2,
      selectedConnector: instance.driver,
      schemaName: driverName,
      connectorInstanceName: instance.name ?? null,
    };
    window.history.pushState(state, "", "");
    window.dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  async function handleBlankSQL() {
    resetModal();
    await createResourceAndNavigate(ResourceKind.Model);
  }

  /**
   * Navigate back to tile selection (step 1).
   * Used by the reactive back-intercept and as the onBack for AddDataForm.
   */
  function goBackToTileSelection() {
    const state = {
      modal: "model" as const,
      step: 1,
      selectedConnector: null,
      schemaName: null,
      connectorInstanceName: null,
    };
    window.history.pushState(state, "", "");
    window.dispatchEvent(new PopStateEvent("popstate", { state }));
    isSubmittingForm = false;
    resetConnectorStep();
  }

  function resetModal() {
    const state = {
      modal: "model" as const,
      step: 0,
      selectedConnector: null,
      schemaName: null,
      connectorInstanceName: null,
    };
    window.history.pushState(state, "", "");
    window.dispatchEvent(new PopStateEvent("popstate", { state }));
    isSubmittingForm = false;
    resetConnectorStep();
  }

  function resetModalQuietly() {
    step = 0;
    selectedConnector = null;
    selectedSchemaName = null;
    isSubmittingForm = false;
    resetConnectorStep();
  }

  async function onCancelDialog() {
    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceCancel,
      BehaviourEventMedium.Button,
      getScreenNameFromPage(),
      MetricsEventSpace.Modal,
    );

    resetModal();
  }

  $: isConnectorType =
    selectedConnector?.implementsObjectStore ||
    selectedConnector?.implementsOlap ||
    selectedConnector?.implementsSqlStore ||
    (selectedConnector?.implementsWarehouse &&
      selectedConnector?.name !== "salesforce") ||
    isMultiStepConnectorSchema(
      getConnectorSchema(selectedSchemaName ?? selectedConnector?.name ?? ""),
    );
</script>

{#if step >= 1}
  <Dialog.Root
    open
    onOpenChange={async (open) => {
      if (!open) {
        await onCancelDialog();
      }
    }}
    closeOnEscape={!isSubmittingForm}
    closeOnOutsideClick={!isSubmittingForm}
  >
    <Dialog.Content
      class={cn(
        "overflow-hidden",
        formWidthClass,
        step === 2 ? "p-0 gap-0" : "p-6 gap-4",
      )}
      noClose={step === 1}
    >
      {#if step === 1}
        <Dialog.Title>Add a Model</Dialog.Title>
        <section>
          <div
            class="connector-grid grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-x-4 gap-y-2"
          >
            <!-- Existing connector instances -->
            {#each existingInstances as instance (instance.name)}
              {#if instance?.driver?.name}
                <button
                  id="instance-{instance.name}"
                  on:click={() => goToExistingConnector(instance)}
                  class="connector-tile-button size-full"
                >
                  <div class="connector-wrapper px-6 py-4">
                    <div class="flex flex-col items-center gap-1">
                      <svelte:component
                        this={ICONS[instance.driver?.name]}
                        size="32px"
                      />
                    </div>
                  </div>
                </button>
              {/if}
            {/each}

            <!-- Uncovered model connector tiles -->
            {#each uncoveredModelConnectors as connector (connector.name)}
              <button
                id={connector.name}
                on:click={() => goToConnectorForm(connector)}
                class="connector-tile-button size-full"
              >
                <div class="connector-wrapper px-6 py-4">
                  <svelte:component this={ICONS[connector.name]} />
                </div>
              </button>
            {/each}

            <!-- Blank SQL tile -->
            <button
              id="blank-sql"
              on:click={handleBlankSQL}
              class="connector-tile-button size-full"
            >
              <div class="connector-wrapper px-6 py-4">
                <div class="flex flex-col items-center gap-1">
                  <FileIcon size="32px" class="text-fg-secondary" />
                  <span class="text-xs text-fg-secondary">Blank SQL</span>
                </div>
              </div>
            </button>
          </div>
        </section>
      {/if}

      {#if step === 2 && pendingConnectorName && !selectedConnector}
        <div class="p-6 flex items-center justify-center">
          <span class="text-fg-secondary">Loading...</span>
        </div>
      {:else if step === 2 && selectedConnector && selectedSchemaName}
        {@const schema = getConnectorSchema(selectedSchemaName)}
        {@const displayIcon =
          connectorIconMapping[selectedSchemaName] ??
          connectorIconMapping[selectedConnector.name ?? ""]}
        {@const displayName = schema?.title ?? selectedConnector.displayName}
        <Dialog.Title class="p-4 border-b border-gray-200">
          <div class="flex items-center gap-[6px]">
            {#if displayIcon}
              <svelte:component this={displayIcon} size="18px" />
            {/if}
            <span class="text-lg leading-none font-semibold">{displayName}</span
            >
          </div>
        </Dialog.Title>

        {#if selectedConnector.name === "local_file"}
          <LocalSourceUpload
            onClose={resetModal}
            onBack={goBackToTileSelection}
          />
        {:else if selectedConnector.name}
          <AddDataForm
            connector={selectedConnector}
            schemaName={selectedSchemaName}
            formType={isConnectorType ? "connector" : "source"}
            {connectorInstanceName}
            onClose={resetModal}
            onCloseAfterNavigation={resetModalQuietly}
            onBack={goBackToTileSelection}
            bind:isSubmitting={isSubmittingForm}
          />
        {/if}
      {/if}
    </Dialog.Content>
  </Dialog.Root>
{/if}

<style lang="postcss">
  section {
    @apply flex flex-col gap-y-3;
  }

  .connector-tile-button {
    aspect-ratio: 2/1;
    @apply basis-40;
    @apply border border-gray-300 rounded;
    @apply cursor-pointer overflow-hidden;
  }

  .connector-wrapper {
    @apply py-3 px-7 size-full;
    @apply flex items-center justify-center;
  }

  .connector-tile-button:hover {
    @apply bg-gray-100;
  }
</style>
