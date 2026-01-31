<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { File } from "lucide-svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { Database, Folder, PlusCircleIcon } from "lucide-svelte";
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
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "../connectors/selectors";
  import { directoryState } from "../file-explorer/directory-store";
  import { createResourceAndNavigate } from "../file-explorer/new-files";
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
    <DropdownMenu.Item
      aria-label="Add Model"
      class="flex gap-x-2"
      disabled={!isModelingSupported}
      on:click={() => createResourceAndNavigate(ResourceKind.Model)}
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
    </DropdownMenu.Item>
    <DropdownMenu.Item
      aria-label="Add Metrics View"
      class="flex gap-x-2"
      on:click={() => createResourceAndNavigate(ResourceKind.MetricsView)}
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
      on:click={() => {
        if (metricsViews.length === 1) {
          void createResourceAndNavigate(
            ResourceKind.Explore,
            metricsViews.pop(),
          );
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
      on:click={() => createResourceAndNavigate(ResourceKind.Canvas)}
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
      <DropdownMenu.SubContent class="w-[240px]">
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
          on:click={() => createResourceAndNavigate(ResourceKind.API)}
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
          on:click={() => createResourceAndNavigate(ResourceKind.Theme)}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.Theme]}
            size="16px"
          />
          Theme
        </DropdownMenu.Item>
        <!-- Temporarily hide Report and Alert options -->
        <!-- <DropdownMenu.Item class="flex gap-x-2" on:click={() => createResourceAndNavigate(ResourceKind.Report)}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Report]}
              className="text-fg-primary"
              size="16px"
            />
            Report
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={() => createResourceAndNavigate(ResourceKind.Alert)}>
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

<CreateExploreDialog bind:open={showExploreDialog} {metricsViews} />

<GenerateSampleData type="modal" bind:open={generateDataDialog} />
