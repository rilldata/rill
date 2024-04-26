<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { Folder, PlusCircleIcon } from "lucide-svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import File from "../../components/icons/File.svelte";
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
  import { featureFlags } from "../feature-flags";
  import { directoryState } from "../file-explorer/directory-store";
  import { handleEntityCreate } from "../file-explorer/new-files";
  import { addSourceModal } from "../sources/modal/add-source-visibility";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../tables/selectors";
  import { removeLeadingSlash } from "./entity-mappers";
  import {
    useDirectoryNamesInDirectory,
    useFileNamesInDirectory,
  } from "./file-selectors";
  import { getName } from "./name-utils";
  import { resourceIconMapping } from "./resource-icon-mapping";
  import { ResourceKind } from "./resource-selectors";

  let active = false;

  const createFile = createRuntimeServicePutFile();
  const createFolder = createRuntimeServiceCreateDirectory();
  const { customDashboards } = featureFlags;

  $: instanceId = $runtime.instanceId;
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

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);

  async function wrapNavigation(toPath: string | undefined) {
    if (!toPath) return;
    const previousScreenName = getScreenNameFromPage();
    await goto(toPath);
    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.Navigate,
      BehaviourEventMedium.Button,
      previousScreenName,
      MetricsEventSpace.LeftPanel,
    );
  }

  /**
   * Open the add source modal
   */
  async function handleAddSource() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      getScreenNameFromPage(),
      MetricsEventSpace.LeftPanel,
    );
  }

  /**
   * Put an example Model file in the `models` directory
   */
  async function handleAddModel() {
    const newRoute = await handleEntityCreate(ResourceKind.Model);
    await wrapNavigation(newRoute);
  }

  /**
   * Put an example Dashboard file in the `dashboards` directory
   */
  async function handleAddDashboard() {
    const newRoute = await handleEntityCreate(ResourceKind.MetricsView);
    await wrapNavigation(newRoute);
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
      path: path,
      data: {
        create: true,
        createOnly: true,
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
      path: path,
      data: {
        blob: undefined,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/${path}`);
  }

  /**
   * Put an example API file in the `apis` directory
   */
  // async function handleAddAPI() {
  //   const newRoute = await handleEntityCreate(ResourceKind.API);
  //   if (newRoute) await goto(newRoute);
  // }

  /**
   * Put an example Chart file in the `charts` directory
   */
  async function handleAddChart() {
    const newRoute = await handleEntityCreate(ResourceKind.Chart);
    await wrapNavigation(newRoute);
  }

  /**
   * Put an example Custom Dashboard file in the `custom-dashbaords` directory
   */
  async function handleAddCustomDashboard() {
    const newRoute = await handleEntityCreate(ResourceKind.Dashboard);
    await wrapNavigation(newRoute);
  }

  /**
   * Put an example Theme file in the `themes` directory
   */
  async function handleAddTheme() {
    const newRoute = await handleEntityCreate(ResourceKind.Theme);
    await wrapNavigation(newRoute);
  }

  /**
   * Put an example Report file in the `reports` directory
   */
  // async function handleAddReport() {
  //   const newRoute = await handleEntityCreate(ResourceKind.Report);
  //   if (newRoute) await goto(newRoute);
  // }

  /**
   * Put an example Alert file in the `alerts` directory
   */
  // async function handleAddAlert() {
  //   const newRoute = await handleEntityCreate(ResourceKind.Alert);
  //   if (newRoute) await goto(newRoute);
  // }
</script>

<div class="p-2">
  <DropdownMenu.Root bind:open={active}>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        {...builder}
        aria-label="Add Asset"
        class="add-asset-button"
        class:open={active}
        use:builder.action
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          Add
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon size="10px" />
          </span>
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" class="w-[240px]">
      {#if $isModelingSupportedForCurrentOlapDriver.data}
        <DropdownMenu.Item
          aria-label="Add Source"
          class="flex gap-x-2"
          on:click={handleAddSource}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.Source]}
            className="text-gray-900"
            size="16px"
          />
          Source
        </DropdownMenu.Item>
        <DropdownMenu.Item
          aria-label="Add Model"
          class="flex gap-x-2"
          on:click={handleAddModel}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.Model]}
            className="text-gray-900"
            size="16px"
          />
          Model
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        aria-label="Add Dashboard"
        class="flex gap-x-2"
        on:click={handleAddDashboard}
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.MetricsView]}
          className="text-gray-900"
          size="16px"
        />
        Dashboard
      </DropdownMenu.Item>
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger>More</DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent class="w-[240px]">
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddFolder}>
            <Folder size="16px" /> Folder
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddBlankFile}>
            <File size="16px" /> Blank file
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
          <!-- Temporarily hide API option -->
          <!-- <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAPI}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.API]}
              className="text-gray-900"
              size="16px"
            />
            API
            <DropdownMenu.Separator />
          </DropdownMenu.Item> -->
          {#if $customDashboards}
            <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddChart}>
              <svelte:component
                this={resourceIconMapping[ResourceKind.Chart]}
                className="text-gray-900"
                size="16px"
              />
              Chart
            </DropdownMenu.Item>
            <DropdownMenu.Item
              class="flex gap-x-2"
              on:click={handleAddCustomDashboard}
            >
              <svelte:component
                this={resourceIconMapping[ResourceKind.Dashboard]}
                className="text-gray-900"
                size="16px"
              />
              Custom Dashboard
            </DropdownMenu.Item>
          {/if}
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddTheme}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Theme]}
              className="text-gray-900"
              size="16px"
            />
            Theme
          </DropdownMenu.Item>
          <!-- Temporarily hide Report and Alert options -->
          <!-- <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddReport}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Report]}
              className="text-gray-900"
              size="16px"
            />
            Report
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAlert}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Alert]}
              className="text-gray-900"
              size="16px"
            />
            Alert
          </DropdownMenu.Item> -->
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

<style lang="postcss">
  .add-asset-button {
    @apply w-full h-7 p-2 rounded-sm;
    @apply flex gap-x-2 items-center justify-center;
    @apply text-primary-700 font-medium bg-primary-50;
  }

  .add-asset-button:hover {
    @apply text-primary-800 bg-primary-100;
  }

  .add-asset-button.open {
    @apply text-primary-900 bg-primary-200;
  }
</style>
