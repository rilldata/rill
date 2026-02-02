<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { File } from "lucide-svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { Database, Folder, Globe, PlusCircleIcon } from "lucide-svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
  import {
    createRuntimeServiceCreateDirectory,
    createRuntimeServicePutFile,
    createRuntimeServiceAnalyzeConnectors,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "../connectors/selectors";
  import { directoryState } from "../file-explorer/directory-store";
  import { createResourceFile } from "../file-explorer/new-files";
  import { addSourceModal } from "../sources/modal/add-source-visibility";
  import CreateExploreDialog from "./CreateExploreDialog.svelte";
  import { removeLeadingSlash } from "./entity-mappers";
  import {
    useDirectoryNamesInDirectory,
    useFileNamesInDirectory,
  } from "./file-selectors";
  import { getName } from "./name-utils";
  import { resourceIconMapping } from "./resource-icon-mapping";
  import { ResourceKind, useFilteredResources } from "./resource-selectors";
  import { connectorIconMapping } from "../connectors/connector-icon-mapping";
  import { getConnectorIconKey } from "../connectors/connectors-utils";
  import GenerateSampleData from "@rilldata/web-common/features/sample-data/GenerateSampleData.svelte";
  import { Wand } from "lucide-svelte";

  let active = false;
  let showExploreDialog = false;
  let generateDataDialog = false;

  const createFile = createRuntimeServicePutFile();
  const createFolder = createRuntimeServiceCreateDirectory();

  $: ({ instanceId } = $runtime);

  $: currentFile = $page.params.file;
  $: currentDirectory = currentFile
    ? currentFile.split("/").slice(0, -1).join("/")
    : "";

  $: currentDirectoryFileNamesQuery = useFileNamesInDirectory(
    instanceId,
    currentDirectory,
  );
  $: currentDirectoryDirectoryNamesQuery = useDirectoryNamesInDirectory(
    instanceId,
    currentDirectory,
  );

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(instanceId);
  $: isModelingSupported = $isModelingSupportedForDefaultOlapDriver.data;

  $: metricsViewQuery = useFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
  );

  $: metricsViews = $metricsViewQuery?.data ?? [];

  // Get connector resources to check for reconcile errors
  $: connectorResourcesQuery = useFilteredResources(
    instanceId,
    ResourceKind.Connector,
  );
  // Create a set of connector names that have reconcile errors
  $: connectorErrors = new Set(
    ($connectorResourcesQuery?.data ?? [])
      .filter((r) => r.meta?.reconcileError)
      .map((r) => r.meta?.name?.name),
  );

  // Connectors for import data modal (object stores, warehouses, SQL stores)
  $: connectors = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      select: (data) => {
        if (!data?.connectors) return;
        const filtered = data.connectors
          .filter(
            (c) =>
              c?.driver?.implementsObjectStore ||
              c?.driver?.implementsWarehouse ||
              c?.driver?.implementsSqlStore,
          )
          .sort((a, b) => (a?.name as string).localeCompare(b?.name as string));
        return { connectors: filtered };
      },
    },
  });
  $: connectorsData = $connectors.data?.connectors ?? [];

  async function wrapNavigation(toPath: string | undefined) {
    if (!toPath) return;
    const previousScreenName = getScreenNameFromPage();
    await goto(`/files${toPath}`);
    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.Navigate,
      BehaviourEventMedium.Button,
      previousScreenName,
      MetricsEventSpace.LeftPanel,
    );
  }

  /**
   * Open the Add Data modal
   */
  async function handleAddData() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      getScreenNameFromPage(),
      MetricsEventSpace.LeftPanel,
    );
  }

  async function handleAddResource(resourceKind: ResourceKind) {
    const newFilePath = await createResourceFile(resourceKind);
    await wrapNavigation(newFilePath);
  }

  /**
   * Put a folder in the current directory
   */
  async function handleAddFolder() {
    const nextFolderName = getName(
      "untitled_folder",
      $currentDirectoryDirectoryNamesQuery?.data ?? [],
    );
    const path =
      currentDirectory !== ""
        ? `${removeLeadingSlash(currentDirectory)}/${nextFolderName}`
        : nextFolderName;

    await $createFolder.mutateAsync({
      instanceId: instanceId,
      data: {
        path: path,
      },
    });

    // Expand the directory to show the new folder
    const pathWithLeadingSlash = `/${path}`;
    directoryState.expand(pathWithLeadingSlash);
  }

  /**
   * Put a blank file in the current directory
   */
  async function handleAddBlankFile() {
    const nextFileName = getName(
      "untitled_file",
      $currentDirectoryFileNamesQuery?.data ?? [],
    );

    const path =
      currentDirectory !== ""
        ? `${removeLeadingSlash(currentDirectory)}/${nextFileName}`
        : nextFileName;

    await $createFile.mutateAsync({
      instanceId: instanceId,
      data: {
        path,
        blob: undefined,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/${path}`);
  }
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <Button
      builders={[builder]}
      label="Add Asset"
      class="w-full"
      type="secondary"
      selected={active}
    >
      <PlusCircleIcon size="14px" />
      <div class="flex gap-x-1 items-center">
        Add
        <span class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="10px" />
        </span>
      </div>
    </Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class={`w-[${
      !isModelingSupported || metricsViews.length === 0 ? "280px" : "240px"
    }]`}
  >
    <DropdownMenu.Item
      aria-label="Add Data"
      class="flex gap-x-2"
      on:click={handleAddData}
    >
      <svelte:component this={Database} color="#C026D3" size="16px" />
      Data
    </DropdownMenu.Item>
    <DropdownMenu.Sub>
      <DropdownMenu.SubTrigger
        class="flex gap-x-2"
        disabled={!isModelingSupported}
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.Model]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          Model
          {#if !isModelingSupported}
            <span class="text-fg-secondary text-xs">
              Requires a supported OLAP driver
            </span>
          {/if}
        </div>
      </DropdownMenu.SubTrigger>
      <DropdownMenu.SubContent align="start" sideOffset={10} class="w-[240px]">
        {#if connectorsData.length > 0 && isModelingSupported}
          {#each connectorsData as connector (connector.name)}
            {#if connector?.driver?.name}
              {@const hasError = connectorErrors.has(connector.name)}
              <DropdownMenu.Item
                class="flex gap-x-2 {hasError ? 'text-red-600' : ''}"
                disabled={hasError}
                on:click={() => {
                  addSourceModal.open(connector.driver?.name, connector.name);
                }}
              >
                <svelte:component
                  this={connectorIconMapping[getConnectorIconKey(connector)]}
                  size="16px"
                />
                {connector.driver?.displayName ?? connector.name} ({connector.name})
              </DropdownMenu.Item>
            {/if}
          {/each}
          <DropdownMenu.Separator />
        {/if}
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => {
            addSourceModal.open("https");
          }}
        >
          <Globe size="16px" />
          Public Data (https)
        </DropdownMenu.Item>
        <DropdownMenu.Item
          aria-label="Blank file"
          class="flex gap-x-2"
          disabled={!isModelingSupported}
          on:click={() => handleAddResource(ResourceKind.Model)}
        >
          <File size="16px" />
          <div class="flex flex-col items-start">
            Blank file
            {#if !isModelingSupported}
              <span class="text-fg-secondary text-xs">
                Requires a supported OLAP driver
              </span>
            {/if}
          </div>
        </DropdownMenu.Item>
      </DropdownMenu.SubContent>
    </DropdownMenu.Sub>
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      aria-label="Add Metrics View"
      class="flex gap-x-2"
      on:click={() => handleAddResource(ResourceKind.MetricsView)}
    >
      <svelte:component
        this={resourceIconMapping[ResourceKind.MetricsView]}
        size="16px"
      />
      Metrics view
    </DropdownMenu.Item>
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      aria-label="Add Explore Dashboard"
      class="flex gap-x-2"
      disabled={metricsViews.length === 0}
      on:click={async () => {
        if (metricsViews.length === 1) {
          const newFilePath = await createResourceFile(
            ResourceKind.Explore,
            metricsViews.pop(),
          );
          await wrapNavigation(newFilePath);
        } else {
          showExploreDialog = true;
        }
      }}
    >
      <div class="flex gap-x-2 items-center">
        <svelte:component
          this={resourceIconMapping[ResourceKind.Explore]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          Explore dashboard
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              Requires a metrics view
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>

    <DropdownMenu.Item
      class="flex items-center justify-between gap-x-2"
      on:click={async () => {
        const newFilePath = await createResourceFile(ResourceKind.Canvas);
        await wrapNavigation(newFilePath);
      }}
      disabled={metricsViews.length === 0}
    >
      <div class="flex gap-x-2 items-center">
        <svelte:component
          this={resourceIconMapping[ResourceKind.Canvas]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          Canvas dashboard
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              Requires a metrics view
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>
    <DropdownMenu.Separator />
    <DropdownMenu.Sub>
      <DropdownMenu.SubTrigger>More</DropdownMenu.SubTrigger>
      <DropdownMenu.SubContent align="start" sideOffset={10} class="w-[240px]">
        <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddFolder}>
          <Folder size="14px" class="stroke-icon-muted" /> Folder
        </DropdownMenu.Item>
        <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddBlankFile}>
          <File size="14px" class="stroke-icon-muted" /> Blank file
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => (generateDataDialog = true)}
        >
          <Wand size="14px" class="stroke-accent-primary-action" /> Generate data
          using AI (beta)
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => handleAddResource(ResourceKind.API)}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.API]}
            size="16px"
          />
          API
          <DropdownMenu.Separator />
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => handleAddResource(ResourceKind.Theme)}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.Theme]}
            size="16px"
          />
          Theme
        </DropdownMenu.Item>
        <!-- Temporarily hide Report and Alert options -->
        <!-- <DropdownMenu.Item class="flex gap-x-2" on:click={() => handleAddResource(ResourceKind.Report)}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Report]}
              className="text-fg-primary"
              size="16px"
            />
            Report
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={() => handleAddResource(ResourceKind.Alert)}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Alert]}
              className="text-fg-primary"
              size="16px"
            />
            Alert
          </DropdownMenu.Item> -->
      </DropdownMenu.SubContent>
    </DropdownMenu.Sub>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<CreateExploreDialog
  {wrapNavigation}
  bind:open={showExploreDialog}
  {metricsViews}
/>

<GenerateSampleData initializeProject={false} bind:open={generateDataDialog} />
