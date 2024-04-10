<script lang="ts">
  import { goto } from "$app/navigation";
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
  import { createRuntimeServicePutFileAndReconcile } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useDashboardFileNames } from "../dashboards/selectors";
  import {
    NEW_ALERT_FILE_CONTENT,
    NEW_API_FILE_CONTENT,
    NEW_CHART_FILE_CONTENT,
    NEW_MODEL_FILE_CONTENT,
    NEW_REPORT_FILE_CONTENT,
    NEW_THEME_FILE_CONTENT,
  } from "../file-explorer/new-files";
  import { useModelFileNames } from "../models/selectors";
  import { addSourceModal } from "../sources/modal/add-source-visibility";
  import { getName } from "./name-utils";
  import { resourceIconMapping } from "./resource-icon-mapping";
  import { ResourceKind } from "./resource-selectors";

  const createFile = createRuntimeServicePutFileAndReconcile();

  $: instanceId = $runtime.instanceId;

  // TODO: we should only fetch the existing names when needed
  $: useModelNames = useModelFileNames(instanceId);
  $: dashboardNames = useDashboardFileNames(instanceId);

  // TODO: get current directory
  $: currentDirectory = "dir-1";

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
    const newModelName = getName("model", $useModelNames?.data ?? []);

    void $createFile.mutateAsync({
      data: {
        instanceId,
        path: `models/${newModelName}.sql`,
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
    const newDashboardName = getName("dashboard", $dashboardNames?.data ?? []);

    void $createFile.mutateAsync({
      data: {
        instanceId,
        path: `dashboards/${newDashboardName}.yaml`,
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
  function handleAddFolder() {
    console.log("Add folder");
  }

  /**
   * Put a blank file in the current directory
   */
  async function handleAddBlankFile() {
    const nextFileName = getName("file", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `${currentDirectory}/${nextFileName}`,
        blob: undefined,
        create: true,
        createOnly: true,
        strict: false,
      },
    });

    await goto(`/files/${currentDirectory}/${nextFileName}`);
  }

  /**
   * Put an example API file in the `apis` directory
   */
  async function handleAddAPI() {
    const nextFileName = getName("api", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `apis/${nextFileName}.yaml`,
        blob: NEW_API_FILE_CONTENT,
        create: true,
        createOnly: true,
        strict: false,
      },
    });

    await goto(`/files/apis/${nextFileName}.yaml`);
  }

  /**
   * Put an example Chart file in the `charts` directory
   */
  async function handleAddChart() {
    const nextFileName = getName("chart", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `charts/${nextFileName}.yaml`,
        blob: NEW_CHART_FILE_CONTENT,
        create: true,
        createOnly: true,
        strict: false,
      },
    });

    await goto(`/files/charts/${nextFileName}.yaml`);
  }

  /**
   * Put an example Theme file in the `themes` directory
   */
  async function handleAddTheme() {
    const nextFileName = getName("theme", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `themes/${nextFileName}.yaml`,
        blob: NEW_THEME_FILE_CONTENT,
        create: true,
        createOnly: true,
        strict: false,
      },
    });

    await goto(`/files/themes/${nextFileName}.yaml`);
  }

  /**
   * Put an example Report file in the `reports` directory
   */
  async function handleAddReport() {
    const nextFileName = getName("report", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `reports/${nextFileName}.yaml`,
        blob: NEW_REPORT_FILE_CONTENT,
        create: true,
        createOnly: true,
        strict: false,
      },
    });

    await goto(`/files/reports/${nextFileName}.yaml`);
  }

  /**
   * Put an example Alert file in the `alerts` directory
   */
  async function handleAddAlert() {
    const nextFileName = getName("alert", []);

    void $createFile.mutateAsync({
      data: {
        instanceId: instanceId,
        path: `alerts/${nextFileName}.yaml`,
        blob: NEW_ALERT_FILE_CONTENT,
        create: true,
        createOnly: true,
        strict: false,
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
        use:builder.action
        class:open
        class="p-2 bg-primary-50 hover:bg-primary-100 text-primary-700 hover:text-primary-800 w-full flex gap-x-2 items-center font-medium h-7 rounded-sm justify-center"
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          Add
          <CaretDownIcon size="10px" />
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-[240px]" align="start">
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddSource}>
        <svelte:component
          this={resourceIconMapping[ResourceKind.Source]}
          size="16px"
          className="text-gray-900"
        />
        Source
      </DropdownMenu.Item>
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddModel}>
        <svelte:component
          this={resourceIconMapping[ResourceKind.Model]}
          size="16px"
          className="text-gray-900"
        />
        Model
      </DropdownMenu.Item>
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddDashboard}>
        <svelte:component
          this={resourceIconMapping[ResourceKind.MetricsView]}
          size="16px"
          className="text-gray-900"
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
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAPI}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.API]}
              size="16px"
              className="text-gray-900"
            />
            API
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddChart}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Chart]}
              size="16px"
              className="text-gray-900"
            />
            Chart
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddTheme}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Theme]}
              size="16px"
              className="text-gray-900"
            />
            Theme
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddReport}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Report]}
              size="16px"
              className="text-gray-900"
            />
            Report
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAlert}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Alert]}
              size="16px"
              className="text-gray-900"
            />
            Alert
          </DropdownMenu.Item>
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
