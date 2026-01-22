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
  import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "../../connectors/selectors";
  import { duplicateSourceName } from "../sources-store";
  import AddDataForm from "./AddDataForm.svelte";
  import DuplicateSource from "./DuplicateSource.svelte";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import RequestConnectorForm from "./RequestConnectorForm.svelte";
  import {
    connectors,
    getConnectorSchema,
    type ConnectorInfo,
  } from "./connector-schemas";
  import { ICONS } from "./icons";
  import { resetConnectorStep } from "./connectorStepStore";

  import type { ClickHouseConnectorType } from "./constants";

  let step = 0;
  let selectedConnector: null | V1ConnectorDriver = null;
  let requestConnector = false;
  let isSubmittingForm = false;
  let initialClickhouseType: ClickHouseConnectorType | undefined = undefined;

  // Filter connectors by category from JSON schemas
  $: sourceConnectors = connectors.filter((c) => c.category !== "olap");
  $: olapConnectors = connectors.filter((c) => c.category === "olap");

  /**
   * Convert a ConnectorInfo (from schema) to a V1ConnectorDriver-compatible object.
   * Derives implements* flags from the schema's x-category.
   */
  function toConnectorDriver(info: ConnectorInfo): V1ConnectorDriver {
    const schema = getConnectorSchema(info.name);
    const category = schema?.["x-category"];

    return {
      name: info.name,
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
      step = e.state?.step ?? 0;
      requestConnector = e.state?.requestConnector ?? false;
      selectedConnector = e.state?.selectedConnector ?? null;
      initialClickhouseType = e.state?.initialClickhouseType ?? undefined;
    }
    window.addEventListener("popstate", listen);

    return () => {
      window.removeEventListener("popstate", listen);
    };
  });

  function goToConnectorForm(connectorInfo: ConnectorInfo) {
    // Reset multi-step state (auth selection, connector config) when switching connectors.
    resetConnectorStep();

    // Handle ClickHouse Cloud - use the actual "clickhouse" connector but pre-select the type
    let actualConnectorInfo = connectorInfo;
    let clickhouseType: ClickHouseConnectorType | undefined = undefined;

    if (connectorInfo.name === "clickhousecloud") {
      // Use the clickhouse schema but mark it as ClickHouse Cloud
      actualConnectorInfo = {
        ...connectorInfo,
        name: "clickhouse",
      };
      clickhouseType = "clickhouse-cloud";
    }

    const state = {
      step: 2,
      selectedConnector: toConnectorDriver(actualConnectorInfo),
      requestConnector: false,
      initialClickhouseType: clickhouseType,
    };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  function goToRequestConnector() {
    const state = { step: 2, requestConnector: true };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  function back() {
    // Try to go back in browser history
    if (window.history.length > 1) {
      window.history.back();
    } else {
      // If no history to go back to, close the modal
      resetModal();
    }
  }

  function resetModal() {
    const state = {
      step: 0,
      selectedConnector: null,
      requestConnector: false,
      initialClickhouseType: undefined,
    };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state: state }));
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

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver($runtime.instanceId);
  $: isModelingSupported = $isModelingSupportedForDefaultOlapDriver.data;

  // FIXME: excluding salesforce until we implement the table discovery APIs
  $: isConnectorType =
    selectedConnector?.implementsObjectStore ||
    selectedConnector?.implementsOlap ||
    selectedConnector?.implementsSqlStore ||
    (selectedConnector?.implementsWarehouse &&
      selectedConnector?.name !== "salesforce");
</script>

{#if step >= 1 || $duplicateSourceName}
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
        "overflow-hidden max-w-4xl",
        step === 2 ? "p-0 gap-0" : "p-6 gap-4",
      )}
      noClose={step === 1}
    >
      {#if step === 1}
        {#if isModelingSupported}
          <Dialog.Title>Add a source</Dialog.Title>
          <section class="mb-1">
            <div
              class="connector-grid grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-x-4 gap-y-2"
            >
              {#each sourceConnectors as connector (connector.name)}
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
            </div>
          </section>
        {/if}
      {/if}

      {#if step === 1}
        <section>
          <Dialog.Title>Connect an OLAP engine</Dialog.Title>

          <div
            class="connector-grid grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-x-4 gap-y-2"
          >
            {#each olapConnectors as connector (connector.name)}
              <button
                id={connector.name}
                class="connector-tile-button size-full"
                on:click={() => goToConnectorForm(connector)}
              >
                <div class="connector-wrapper px-6 py-4">
                  <svelte:component this={ICONS[connector.name]} />
                </div>
              </button>
            {/each}
          </div>
        </section>

        <div class="text-fg-secondary">
          Don't see what you're looking for?
          <button
            class="text-primary-500 hover:text-primary-600 font-medium"
            on:click={goToRequestConnector}
          >
            Request a new connector
          </button>
        </div>
      {/if}

      {#if step === 2 && selectedConnector}
        {@const isClickhouseCloud =
          initialClickhouseType === "clickhouse-cloud"}
        {@const displayIcon = isClickhouseCloud
          ? connectorIconMapping["clickhousecloud"]
          : connectorIconMapping[selectedConnector.name ?? ""]}
        {@const displayName = isClickhouseCloud
          ? "ClickHouse Cloud"
          : selectedConnector.displayName}
        <Dialog.Title class="p-4 border-b border-gray-200">
          {#if $duplicateSourceName !== null}
            Duplicate source
          {:else}
            <div class="flex items-center gap-[6px]">
              {#if displayIcon}
                <svelte:component this={displayIcon} size="18px" />
              {/if}
              <span class="text-lg leading-none font-semibold"
                >{displayName}</span
              >
            </div>
          {/if}
        </Dialog.Title>

        {#if $duplicateSourceName !== null}
          <div class="p-6">
            <DuplicateSource onCancel={resetModal} onComplete={resetModal} />
          </div>
        {:else if selectedConnector.name === "local_file"}
          <LocalSourceUpload onClose={resetModal} onBack={back} />
        {:else if selectedConnector.name}
          <AddDataForm
            connector={selectedConnector}
            formType={isConnectorType ? "connector" : "source"}
            onClose={resetModal}
            onBack={back}
            bind:isSubmitting={isSubmittingForm}
            {initialClickhouseType}
          />
        {/if}
      {/if}

      {#if step === 2 && requestConnector}
        <div class="p-6">
          <Dialog.Title>Request a connector</Dialog.Title>
          <RequestConnectorForm onClose={resetModal} onBack={back} />
        </div>
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
