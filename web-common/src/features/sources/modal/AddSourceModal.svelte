<script lang="ts">
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import {
    createRuntimeServiceListConnectors,
    V1ConnectorSpec,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import AmazonS3 from "../../../components/icons/connectors/AmazonS3.svelte";
  import MicrosoftAzureBlobStorage from "../../../components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
  import GoogleBigQuery from "../../../components/icons/connectors/GoogleBigQuery.svelte";
  import GoogleCloudStorage from "../../../components/icons/connectors/GoogleCloudStorage.svelte";
  import Https from "../../../components/icons/connectors/HTTPS.svelte";
  import LocalFile from "../../../components/icons/connectors/LocalFile.svelte";
  import MotherDuck from "../../../components/icons/connectors/MotherDuck.svelte";
  import Postgres from "../../../components/icons/connectors/Postgres.svelte";
  import { appScreen } from "../../../layout/app-store";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import RemoteSourceForm from "./RemoteSourceForm.svelte";
  import RequestConnectorForm from "./RequestConnectorForm.svelte";
  import AmazonAthena from "@rilldata/web-common/components/icons/connectors/AmazonAthena.svelte";

  export let open: boolean;

  let step = 1;
  let selectedConnector: V1ConnectorSpec;
  let requestConnector = false;

  const TAB_ORDER = [
    "gcs",
    "s3",
    "azure",
    // duckdb
    "bigquery",
    "athena",
    "motherduck",
    "postgres",
    "local_file",
    "https",
  ];

  const ICONS = {
    gcs: GoogleCloudStorage,
    s3: AmazonS3,
    azure: MicrosoftAzureBlobStorage,
    // duckdb: DuckDB,
    bigquery: GoogleBigQuery,
    athena: AmazonAthena,
    motherduck: MotherDuck,
    postgres: Postgres,
    local_file: LocalFile,
    https: Https,
  };

  const connectors = createRuntimeServiceListConnectors({
    query: {
      // arrange connectors in the way we would like to display them
      select: (data) => {
        data.connectors =
          data.connectors &&
          data.connectors.sort(
            (a, b) => TAB_ORDER.indexOf(a.name) - TAB_ORDER.indexOf(b.name)
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
    selectedConnector = undefined;
    step = 1;
  }

  const dispatch = createEventDispatcher();

  function onCompleteDialog() {
    resetModal();
    dispatch("close");
  }

  function onCancelDialog() {
    behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceCancel,
      BehaviourEventMedium.Button,
      $appScreen,
      MetricsEventSpace.Modal
    );
    resetModal();
    dispatch("close");
  }
</script>

<!-- This precise width fits exactly 3 connectors per line  -->
<Dialog {open} on:close={onCancelDialog} widthOverride="w-[560px]">
  <div slot="title">
    {#if step === 1}
      Add a source
    {:else if step === 2}
      <h2 class="flex gap-x-1 items-center">
        <button
          on:click={resetModal}
          class="text-gray-500 text-sm font-semibold hover:text-gray-700"
        >
          Add a source
        </button>
        <CaretDownIcon
          size="14px"
          className="transform -rotate-90 text-gray-500"
        />
        <span>
          {#if selectedConnector}
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
    {#if step === 1}
      {#if $connectors.data}
        <div class="grid grid-cols-3 gap-4">
          {#each $connectors.data.connectors as connector}
            <button
              id={connector.name}
              on:click={() => goToConnectorForm(connector)}
              class="w-40 h-20 rounded border border-gray-300 justify-center items-center gap-2.5 inline-flex hover:bg-gray-100 cursor-pointer"
            >
              <svelte:component this={ICONS[connector.name]} />
            </button>
          {/each}
        </div>
      {/if}
      <div>
        Don't see what you're looking for? <button
          on:click={goToRequestConnector}
          class="text-blue-500">Request a new connector</button
        >
      </div>
    {:else if step === 2}
      {#if selectedConnector}
        {#if selectedConnector.name === "local_file"}
          <LocalSourceUpload on:close={onCompleteDialog} />
        {:else}
          <RemoteSourceForm
            connector={selectedConnector}
            on:close={onCompleteDialog}
          />
        {/if}
      {/if}

      {#if requestConnector}
        <RequestConnectorForm on:close={onCompleteDialog} />
      {/if}
    {/if}
  </div>
</Dialog>
