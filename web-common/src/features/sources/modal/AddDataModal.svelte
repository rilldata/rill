<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AmazonAthena from "@rilldata/web-common/components/icons/connectors/AmazonAthena.svelte";
  import AmazonRedshift from "@rilldata/web-common/components/icons/connectors/AmazonRedshift.svelte";
  import MySQL from "@rilldata/web-common/components/icons/connectors/MySQL.svelte";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import {
    createRuntimeServiceListConnectorDrivers,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import AmazonS3 from "../../../components/icons/connectors/AmazonS3.svelte";
  import ApacheDruid from "../../../components/icons/connectors/ApacheDruid.svelte";
  import ApachePinot from "../../../components/icons/connectors/ApachePinot.svelte";
  import ClickHouse from "../../../components/icons/connectors/ClickHouse.svelte";
  import DuckDB from "../../../components/icons/connectors/DuckDB.svelte";
  import GoogleBigQuery from "../../../components/icons/connectors/GoogleBigQuery.svelte";
  import GoogleCloudStorage from "../../../components/icons/connectors/GoogleCloudStorage.svelte";
  import Https from "../../../components/icons/connectors/HTTPS.svelte";
  import LocalFile from "../../../components/icons/connectors/LocalFile.svelte";
  import MicrosoftAzureBlobStorage from "../../../components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
  import MotherDuck from "../../../components/icons/connectors/MotherDuck.svelte";
  import Postgres from "../../../components/icons/connectors/Postgres.svelte";
  import Salesforce from "../../../components/icons/connectors/Salesforce.svelte";
  import Snowflake from "../../../components/icons/connectors/Snowflake.svelte";
  import SQLite from "../../../components/icons/connectors/SQLite.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsModelingSupportedForDefaultOlapDriver } from "../../connectors/olap/selectors";
  import { duplicateSourceName } from "../sources-store";
  import AddDataForm from "./AddDataForm.svelte";
  import DuplicateSource from "./DuplicateSource.svelte";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import RequestConnectorForm from "./RequestConnectorForm.svelte";
  import { connectorIconMapping } from "../../connectors/connector-icon-mapping";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  let step = 0;
  let selectedConnector: null | V1ConnectorDriver = null;
  let requestConnector = false;
  let isSubmittingForm = false;

  const SOURCES = [
    "gcs",
    "s3",
    "azure",
    "bigquery",
    "athena",
    "redshift",
    "duckdb",
    "motherduck",
    "postgres",
    "mysql",
    "sqlite",
    "snowflake",
    "salesforce",
    "local_file",
    "https",
  ];

  const OLAP_CONNECTORS = ["clickhouse", "druid", "pinot"];

  const SORT_ORDER = [...SOURCES, ...OLAP_CONNECTORS];

  const ICONS = {
    gcs: GoogleCloudStorage,
    s3: AmazonS3,
    azure: MicrosoftAzureBlobStorage,
    bigquery: GoogleBigQuery,
    athena: AmazonAthena,
    redshift: AmazonRedshift,
    duckdb: DuckDB,
    motherduck: MotherDuck,
    postgres: Postgres,
    mysql: MySQL,
    sqlite: SQLite,
    snowflake: Snowflake,
    salesforce: Salesforce,
    local_file: LocalFile,
    https: Https,
    clickhouse: ClickHouse,
    druid: ApacheDruid,
    pinot: ApachePinot,
  };

  const connectorsQuery = createRuntimeServiceListConnectorDrivers({
    query: {
      // arrange connectors in the way we would like to display them
      select: (data) => {
        data.connectors =
          data.connectors &&
          data.connectors
            .filter(
              // Only show connectors in SOURCES or OLAP_CONNECTORS
              (a) =>
                a.name &&
                (SOURCES.includes(a.name) || OLAP_CONNECTORS.includes(a.name)),
            )
            .sort(
              // CAST SAFETY: we have filtered out any connectors that
              // don't have a `name` in the previous filter
              (a, b) =>
                SORT_ORDER.indexOf(a.name as string) -
                SORT_ORDER.indexOf(b.name as string),
            );
        return data;
      },
    },
  });

  $: connectors = $connectorsQuery.data?.connectors ?? [];

  onMount(() => {
    function listen(e: PopStateEvent) {
      step = e.state?.step ?? 0;
      requestConnector = e.state?.requestConnector ?? false;
      selectedConnector = e.state?.selectedConnector ?? null;
    }
    window.addEventListener("popstate", listen);

    return () => {
      window.removeEventListener("popstate", listen);
    };
  });

  function goToConnectorForm(connector: V1ConnectorDriver) {
    const state = {
      step: 2,
      selectedConnector: connector,
      requestConnector: false,
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
    window.history.back();
  }

  function handleSubmittingChange(event: CustomEvent) {
    isSubmittingForm = event.detail.submitting;
  }

  function resetModal() {
    const state = { step: 0, selectedConnector: null, requestConnector: false };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state: state }));
    isSubmittingForm = false;
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
        "overflow-hidden",
        step === 2 ? "max-w-4xl p-0 gap-0" : "p-6 gap-4",
      )}
      noClose={step === 1}
    >
      {#if step === 1}
        {#if $isModelingSupportedForDefaultOlapDriver}
          <Dialog.Title>Add a source</Dialog.Title>
          <section class="mb-1">
            <div class="connector-grid">
              {#each connectors.filter((c) => c.name && SOURCES.includes(c.name)) as connector (connector.name)}
                {#if connector.name}
                  <button
                    id={connector.name}
                    on:click={() => goToConnectorForm(connector)}
                    class="connector-tile-button"
                  >
                    <div class="connector-wrapper">
                      <svelte:component this={ICONS[connector.name]} />
                    </div>
                  </button>
                {/if}
              {/each}
            </div>
          </section>
        {/if}
      {/if}

      {#if step === 1}
        <section>
          <Dialog.Title>Connect an OLAP engine</Dialog.Title>

          <div class="connector-grid">
            {#each connectors?.filter((c) => c.name && OLAP_CONNECTORS.includes(c.name)) as connector (connector.name)}
              {#if connector.name}
                <button
                  id={connector.name}
                  class="connector-tile-button"
                  on:click={() => goToConnectorForm(connector)}
                >
                  <div class="connector-wrapper">
                    <svelte:component this={ICONS[connector.name]} />
                  </div>
                </button>
              {/if}
            {/each}
          </div>
        </section>

        <div class="text-slate-500">
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
        <Dialog.Title class="p-4 border-b border-gray-200">
          {#if $duplicateSourceName !== null}
            Duplicate source
          {:else}
            <div class="flex items-center gap-[6px]">
              {#if selectedConnector?.name}
                <svelte:component
                  this={connectorIconMapping[selectedConnector.name]}
                  size="18px"
                />
              {/if}
              <span class="text-lg leading-none font-semibold"
                >{selectedConnector.displayName}</span
              >
            </div>
          {/if}
        </Dialog.Title>

        {#if $duplicateSourceName !== null}
          <div class="p-6">
            <DuplicateSource onCancel={resetModal} onComplete={resetModal} />
          </div>
        {:else if selectedConnector.name === "local_file"}
          <LocalSourceUpload on:close={resetModal} on:back={back} />
        {:else if selectedConnector.name}
          <AddDataForm
            connector={selectedConnector}
            formType={OLAP_CONNECTORS.includes(selectedConnector.name)
              ? "connector"
              : "source"}
            onClose={resetModal}
            onBack={back}
            on:submitting={handleSubmittingChange}
          />
        {/if}
      {/if}

      {#if step === 2 && requestConnector}
        <div class="p-6">
          <Dialog.Title>Request a connector</Dialog.Title>
          <RequestConnectorForm on:close={resetModal} on:back={back} />
        </div>
      {/if}
    </Dialog.Content>
  </Dialog.Root>
{/if}

<style lang="postcss">
  section {
    @apply flex flex-col gap-y-3;
  }

  .connector-grid {
    @apply grid grid-cols-3 gap-4;
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
