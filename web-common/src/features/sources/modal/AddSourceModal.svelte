<script lang="ts">
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import AmazonAthena from "@rilldata/web-common/components/icons/connectors/AmazonAthena.svelte";
  import AmazonRedshift from "@rilldata/web-common/components/icons/connectors/AmazonRedshift.svelte";

  import MySQL from "@rilldata/web-common/components/icons/connectors/MySQL.svelte";
  import {
    createRuntimeServiceListConnectors,
    V1ConnectorSpec,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import AmazonS3 from "../../../components/icons/connectors/AmazonS3.svelte";
  import ClickHouse from "../../../components/icons/connectors/ClickHouse.svelte";
  import DuckDB from "../../../components/icons/connectors/DuckDB.svelte";
  import GoogleBigQuery from "../../../components/icons/connectors/GoogleBigQuery.svelte";
  import GoogleCloudStorage from "../../../components/icons/connectors/GoogleCloudStorage.svelte";
  import Https from "../../../components/icons/connectors/HTTPS.svelte";
  import LocalFile from "../../../components/icons/connectors/LocalFile.svelte";
  import MicrosoftAzureBlobStorage from "../../../components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
  import Postgres from "../../../components/icons/connectors/Postgres.svelte";
  import Salesforce from "../../../components/icons/connectors/Salesforce.svelte";
  import Snowflake from "../../../components/icons/connectors/Snowflake.svelte";
  import SQLite from "../../../components/icons/connectors/SQLite.svelte";
  import { appScreen } from "../../../layout/app-store";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import ClickHouseInstructions from "./ClickHouseInstructions.svelte";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import RemoteSourceForm from "./RemoteSourceForm.svelte";
  import RequestConnectorForm from "./RequestConnectorForm.svelte";
  import { duplicateSourceName } from "../sources-store";
  import DuplicateSource from "./DuplicateSource.svelte";
  import { addSourceModal } from "./add-source-visibility";

  export let step = 1;
  export let selectedConnector: null | V1ConnectorSpec = null;

  let requestConnector = false;

  const TAB_ORDER = [
    "gcs",
    "s3",
    "azure",
    // duckdb
    "bigquery",
    "athena",
    "redshift",
    "duckdb",
    "postgres",
    "mysql",
    "sqlite",
    "snowflake",
    "salesforce",
    "local_file",
    "https",
    "clickhouse",
  ];

  const ICONS = {
    gcs: GoogleCloudStorage,
    s3: AmazonS3,
    azure: MicrosoftAzureBlobStorage,
    // duckdb: DuckDB,
    bigquery: GoogleBigQuery,
    athena: AmazonAthena,
    redshift: AmazonRedshift,
    duckdb: DuckDB,
    postgres: Postgres,
    mysql: MySQL,
    sqlite: SQLite,
    snowflake: Snowflake,
    salesforce: Salesforce,
    local_file: LocalFile,
    https: Https,
    clickhouse: ClickHouse,
  };

  const connectors = createRuntimeServiceListConnectors({
    query: {
      // arrange connectors in the way we would like to display them
      select: (data) => {
        data.connectors =
          data.connectors &&
          data.connectors
            .filter(
              // Only show connectors in TAB_ORDER
              (a) => a.name && TAB_ORDER.indexOf(a.name) >= 0,
            )
            .sort(
              // CAST SAFETY: we have filtered out any connectors that
              // don't have a `name` in the previous filter
              (a, b) =>
                TAB_ORDER.indexOf(a.name as string) -
                TAB_ORDER.indexOf(b.name as string),
            );
        return data;
      },
    },
  });

  function goToConnectorForm(connector: V1ConnectorSpec) {
    selectedConnector = connector;
    step = 2;
  }
  function goToRequestConnector() {
    requestConnector = true;
    step = 2;
  }
  function resetModal() {
    requestConnector = false;
    selectedConnector = null;
    step = 1;
    addSourceModal.close();
  }

  const dispatch = createEventDispatcher();

  function onCompleteDialog() {
    resetModal();
    dispatch("close");
  }

  async function onCancelDialog() {
    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceCancel,
      BehaviourEventMedium.Button,
      $appScreen,
      MetricsEventSpace.Modal,
    );
    resetModal();
    dispatch("close");
  }
</script>

<!-- This precise width fits exactly 3 connectors per line  -->
<Dialog open on:close={onCancelDialog} widthOverride="w-[560px]">
  <div slot="title">
    {#if step === 1}
      Add a source
    {:else if step === 2}
      <h2 class="flex gap-x-1 items-center">
        <span>
          {#if $duplicateSourceName !== null}
            Duplicate source
          {:else if selectedConnector}
            {selectedConnector?.displayName}
          {/if}

          {#if requestConnector}
            Request a connector
          {/if}
        </span>
      </h2>
    {/if}
  </div>
  <div slot="body" class="flex flex-col gap-y-4">
    {#if $duplicateSourceName}
      <DuplicateSource
        on:cancel={onCompleteDialog}
        on:complete={onCompleteDialog}
      />
    {:else if step === 1}
      {#if $connectors.data}
        <div class="grid grid-cols-3 gap-4">
          {#each $connectors?.data?.connectors ?? [] as connector}
            {#if connector.name}
              <button
                id={connector.name}
                on:click={() => goToConnectorForm(connector)}
                class="w-40 h-20 rounded border border-gray-300 justify-center items-center gap-2.5 inline-flex hover:bg-gray-100 cursor-pointer"
              >
                <svelte:component this={ICONS[connector.name]} />
              </button>
            {/if}
          {/each}
        </div>
      {/if}
      <div class="text-slate-500">
        Don't see what you're looking for? <button
          on:click={goToRequestConnector}
          class="text-primary-500 hover:text-primary-600 font-medium"
          >Request a new connector</button
        >
      </div>
    {:else if step === 2}
      {#if selectedConnector}
        {#if selectedConnector.name === "local_file"}
          <LocalSourceUpload on:close={onCompleteDialog} on:back={resetModal} />
        {:else if selectedConnector.name === "clickhouse"}
          <ClickHouseInstructions on:back={resetModal} />
        {:else}
          <RemoteSourceForm
            connector={selectedConnector}
            on:close={onCompleteDialog}
            on:back={resetModal}
          />
        {/if}
      {/if}

      {#if requestConnector}
        <RequestConnectorForm
          on:close={onCompleteDialog}
          on:back={resetModal}
        />
      {/if}
    {/if}
  </div>
</Dialog>
