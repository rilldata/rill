<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Folder, PlusCircleIcon } from "lucide-svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import File from "../../components/icons/File.svelte";
  import { appScreen } from "../../layout/app-store";
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
  import { useAlertFileNames } from "../alerts/selectors";
  import { useAPIFileNames } from "../apis/selectors";
  import { useChartFileNames } from "../charts/selectors";
  import { useDashboardFileNames } from "../dashboards/selectors";
  import { featureFlags } from "../feature-flags";
  import {
    NEW_ALERT_FILE_CONTENT,
    NEW_API_FILE_CONTENT,
    NEW_CHART_FILE_CONTENT,
    NEW_MODEL_FILE_CONTENT,
    NEW_REPORT_FILE_CONTENT,
    NEW_THEME_FILE_CONTENT,
  } from "../file-explorer/new-files";
  import { useModelFileNames } from "../models/selectors";
  import { useReportFileNames } from "../reports/selectors";
  import { addSourceModal } from "../sources/modal/add-source-visibility";
  import { useThemeFileNames } from "../themes/selectors";
  import {
    useDirectoryNamesInDirectory,
    useFileNamesInDirectory,
  } from "./file-selectors";
  import { getName } from "./name-utils";
  import { resourceIconMapping } from "./resource-icon-mapping";
  import { ResourceKind } from "./resource-selectors";

  const createFile = createRuntimeServicePutFile();
  const createFolder = createRuntimeServiceCreateDirectory();
  const { customDashboards } = featureFlags;

  $: instanceId = $runtime.instanceId;
  $: currentFile = $page.params.file;
  $: currentDirectory =
    currentFile && currentFile.split("/").slice(0, -1).join("/");

  // TODO: we should only fetch the existing names when needed
  // TODO: simplify all this
  $: currentDirectoryFileNamesQuery = useFileNamesInDirectory(
    instanceId,
    currentDirectory,
  );
  $: currentDirectoryDirectoryNamesQuery = useDirectoryNamesInDirectory(
    instanceId,
    currentDirectory,
  );
  $: modelFileNamesQuery = useModelFileNames(instanceId);
  $: dashboardFileNamesQuery = useDashboardFileNames(instanceId);
  $: apiFileNamesQuery = useAPIFileNames(instanceId);
  $: chartFileNamesQuery = useChartFileNames(instanceId);
  $: themeFileNamesQuery = useThemeFileNames(instanceId);
  $: reportFileNamesQuery = useReportFileNames(instanceId);
  $: alertFileNamesQuery = useAlertFileNames(instanceId);

  /**
   * Open the add source modal
   */
  async function handleAddSource() {
    addSourceModal.open();

    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceAdd,
      BehaviourEventMedium.Button,
      $appScreen.type,
      MetricsEventSpace.LeftPanel,
    );
  }

  /**
   * Put an example Model file in the `models` directory
   */
  async function handleAddModel() {
    const newModelName = getName("model", $modelFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId,
      path: `models/${newModelName}.sql`,
      data: {
        blob: NEW_MODEL_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/models/${newModelName}.sql`);
  }

  /**
   * Put an example Dashboard file in the `dashboards` directory
   */
  async function handleAddDashboard() {
    const newDashboardName = getName(
      "dashboard",
      $dashboardFileNamesQuery?.data ?? [],
    );

    void $createFile.mutateAsync({
      instanceId,
      path: `dashboards/${newDashboardName}.yaml`,
      data: {
        blob: "",
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/dashboards/${newDashboardName}.yaml`);
  }

  /**
   * Put a folder in the current directory
   */
  async function handleAddFolder() {
    const nextFolderName = getName(
      "untitled_folder",
      $currentDirectoryDirectoryNamesQuery?.data ?? [],
    );

    await $createFolder.mutateAsync({
      instanceId: instanceId,
      path: `${currentDirectory}/${nextFolderName}`,
      data: {
        create: true,
        createOnly: true,
      },
    });
  }

  /**
   * Put a blank file in the current directory
   */
  async function handleAddBlankFile() {
    const nextFileName = getName(
      "untitled_file",
      $currentDirectoryFileNamesQuery?.data ?? [],
    );

    await $createFile.mutateAsync({
      instanceId: instanceId,
      path: `${currentDirectory}/${nextFileName}`,
      data: {
        blob: undefined,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/${currentDirectory}/${nextFileName}`);
  }

  /**
   * Put an example API file in the `apis` directory
   */
  async function handleAddAPI() {
    const nextFileName = getName("api", $apiFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId: instanceId,
      path: `apis/${nextFileName}.yaml`,
      data: {
        blob: NEW_API_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/apis/${nextFileName}.yaml`);
  }

  /**
   * Put an example Chart file in the `charts` directory
   */
  async function handleAddChart() {
    const nextFileName = getName("chart", $chartFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId: instanceId,
      path: `charts/${nextFileName}.yaml`,
      data: {
        blob: NEW_CHART_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/charts/${nextFileName}.yaml`);
  }

  /**
   * Put an example Theme file in the `themes` directory
   */
  async function handleAddTheme() {
    const nextFileName = getName("theme", $themeFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId: instanceId,
      path: `themes/${nextFileName}.yaml`,
      data: {
        blob: NEW_THEME_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/themes/${nextFileName}.yaml`);
  }

  /**
   * Put an example Report file in the `reports` directory
   */
  async function handleAddReport() {
    const nextFileName = getName("report", $reportFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId: instanceId,
      path: `reports/${nextFileName}.yaml`,
      data: {
        blob: NEW_REPORT_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/reports/${nextFileName}.yaml`);
  }

  /**
   * Put an example Alert file in the `alerts` directory
   */
  async function handleAddAlert() {
    const nextFileName = getName("alert", $alertFileNamesQuery?.data ?? []);

    void $createFile.mutateAsync({
      instanceId: instanceId,
      path: `alerts/${nextFileName}.yaml`,
      data: {
        blob: NEW_ALERT_FILE_CONTENT,
        create: true,
        createOnly: true,
      },
    });

    await goto(`/files/alerts/${nextFileName}.yaml`);
  }
</script>

<div class="p-2">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        {...builder}
        aria-label="Add Asset"
        class="p-2 bg-primary-50 hover:bg-primary-100 text-primary-700 hover:text-primary-800 w-full flex gap-x-2 items-center font-medium h-7 rounded-sm justify-center"
        class:open
        use:builder.action
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          Add
          <CaretDownIcon size="10px" />
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" class="w-[240px]">
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
