<script lang="ts">
  import { page } from "$app/stores";
  import {
    Bot,
    Database,
    File,
    Folder,
    GraduationCap,
    PlusCircleIcon,
    Wand,
  } from "lucide-svelte";
  import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
  import Button from "../../../components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry.ts";
  import GenerateSampleData from "@rilldata/web-common/features/sample-data/GenerateSampleData.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes.ts";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../../metrics/service/MetricsTypes.ts";
  import {
    createRuntimeServiceCreateDirectoryMutation,
    createRuntimeServicePutFileMutation,
  } from "../../../runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "../../connectors/selectors.ts";
  import { directoryState } from "../../file-explorer/directory-store.ts";
  import ChartTypesDialog from "@rilldata/web-common/features/custom-viz/examples/ChartTypesDialog.svelte";
  import { createResourceAndNavigate, skillFileTemplate } from "./new-files.ts";
  import AddAiConnectorDialog from "../../connectors/ai/AddAiConnectorDialog.svelte";
  import CreateExploreDialog from "./CreateExploreDialog.svelte";
  import { removeLeadingSlash } from "../entity-mappers.ts";
  import {
    useDirectoryNamesInDirectory,
    useFileNamesInDirectory,
  } from "../file-selectors.ts";
  import { getName } from "../name-utils.ts";
  import { resourceIconMapping } from "../resource-icon-mapping.ts";
  import { ResourceKind, useFilteredResources } from "../resource-selectors.ts";
  import AddModelSubOption from "@rilldata/web-common/features/entity-management/add/AddModelSubOption.svelte";
  import AddDataModal from "@rilldata/web-common/features/add-data/AddDataModal.svelte";
  import AddMetricsViewSubOption from "@rilldata/web-common/features/entity-management/add/AddMetricsViewSubOption.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let active = false;
  let showExploreDialog = false;
  let generateDataDialog = false;
  let showAiConnectorDialog = false;
  let showChartTypesDialog = false;
  let addDataModalOpen = false;
  let addDataConnector = "";
  let addDataTargetResource: ResourceKind | undefined;

  let screenName = MetricsEventScreenName.Home;

  const runtimeClient = useRuntimeClient();

  const createFile = createRuntimeServicePutFileMutation(runtimeClient);
  const createFolder =
    createRuntimeServiceCreateDirectoryMutation(runtimeClient);
  const { developerChat, customComponents } = featureFlags;

  $: currentFile = $page.params.file;
  $: currentDirectory = currentFile
    ? currentFile.split("/").slice(0, -1).join("/")
    : "";

  $: currentDirectoryFileNamesQuery = useFileNamesInDirectory(
    runtimeClient,
    currentDirectory,
  );
  $: currentDirectoryDirectoryNamesQuery = useDirectoryNamesInDirectory(
    runtimeClient,
    currentDirectory,
  );

  // Skills are keyed by name across both supported roots, so a new skill must not collide with either
  $: skillDirectoryNamesQuery = useDirectoryNamesInDirectory(
    runtimeClient,
    "skills",
  );
  $: agentsSkillDirectoryNamesQuery = useDirectoryNamesInDirectory(
    runtimeClient,
    ".agents/skills",
  );

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(runtimeClient);
  $: isModelingSupported = $isModelingSupportedForDefaultOlapDriver.data;

  $: metricsViewQuery = useFilteredResources(
    runtimeClient,
    ResourceKind.MetricsView,
  );

  $: metricsViews = $metricsViewQuery?.data ?? [];

  /**
   * Open the Add Data modal
   */
  function handleAddData() {
    addDataModalOpen = true;
    addDataConnector = "";
    addDataTargetResource = undefined;
    screenName = getScreenNameFromPage();
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
      path: path,
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
      path,
      blob: undefined,
      create: true,
      createOnly: true,
    });

    await navigateToFile(`/${path}`);
  }

  /**
   * Put a skill file (markdown instructions for the AI agents) in the skills directory
   */
  async function handleAddSkill() {
    // Skill names only allow lowercase letters, numbers and hyphens, so we can't use getName, which appends "_N" suffixes
    const existingNames = new Set(
      [
        ...($skillDirectoryNamesQuery?.data ?? []),
        ...($agentsSkillDirectoryNamesQuery?.data ?? []),
      ].map((n) => n.toLowerCase()),
    );
    let name = "my-skill";
    for (let i = 1; existingNames.has(name); i++) {
      name = `my-skill-${i}`;
    }
    const path = `skills/${name}/SKILL.md`;

    await $createFile.mutateAsync({
      path,
      blob: skillFileTemplate(name),
      create: true,
      createOnly: true,
    });

    await navigateToFile(`/${path}`);
  }
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        label={m.add_asset_label()}
        class="w-full"
        type="secondary"
        selected={active}
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          {m.add_asset_add()}
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon size="10px" />
          </span>
        </div>
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class={`w-[${
      !isModelingSupported || metricsViews.length === 0 ? "280px" : "240px"
    }]`}
  >
    <DropdownMenu.Item
      aria-label={m.add_asset_add_data()}
      class="flex gap-x-2"
      onclick={handleAddData}
    >
      <svelte:component this={Database} color="#C026D3" size="16px" />
      {m.add_asset_data()}
    </DropdownMenu.Item>
    <AddModelSubOption
      onSelect={(connector) => {
        addDataModalOpen = true;
        addDataConnector = connector;
        addDataTargetResource = ResourceKind.Model;
      }}
    />
    <AddMetricsViewSubOption
      onSelect={(connector) => {
        addDataModalOpen = true;
        addDataConnector = connector;
        addDataTargetResource = ResourceKind.MetricsView;
      }}
    />
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      aria-label={m.add_asset_add_explore_dashboard()}
      class="flex gap-x-2"
      disabled={metricsViews.length === 0}
      onclick={() => {
        if (metricsViews.length === 1) {
          void createResourceAndNavigate(
            runtimeClient,
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
          {m.add_asset_explore_dashboard()}
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              {m.add_asset_requires_metrics_view()}
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>

    <DropdownMenu.Item
      class="flex items-center justify-between gap-x-2"
      onclick={() =>
        createResourceAndNavigate(runtimeClient, ResourceKind.Canvas)}
      disabled={metricsViews.length === 0}
    >
      <div class="flex gap-x-2 items-center">
        <svelte:component
          this={resourceIconMapping[ResourceKind.Canvas]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          {m.add_asset_canvas_dashboard()}
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              {m.add_asset_requires_metrics_view()}
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>
    {#if $customComponents}
      <DropdownMenu.Item
        class="flex gap-x-2 items-center"
        onclick={() => (showChartTypesDialog = true)}
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.Component]}
          size="16px"
        />
        {m.component_custom_viz()}
      </DropdownMenu.Item>
    {/if}
    <DropdownMenu.Separator />
    <DropdownMenu.Sub>
      <DropdownMenu.SubTrigger>{m.add_asset_more()}</DropdownMenu.SubTrigger>
      <DropdownMenu.SubContent class="w-[240px]">
        <DropdownMenu.Item class="flex gap-x-2" onclick={handleAddFolder}>
          <Folder size="14px" class="stroke-icon-muted" />
          {m.add_asset_folder()}
        </DropdownMenu.Item>
        <DropdownMenu.Item class="flex gap-x-2" onclick={handleAddBlankFile}>
          <File size="14px" class="stroke-icon-muted" />
          {m.add_asset_blank_file()}
        </DropdownMenu.Item>
        {#if $developerChat}
          <DropdownMenu.Item
            class="flex gap-x-2"
            onclick={() => (generateDataDialog = true)}
          >
            <Wand size="14px" class="stroke-accent-primary-action" />
            {m.add_asset_generate_data_with_ai()}
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex gap-x-2"
          onclick={() =>
            createResourceAndNavigate(runtimeClient, ResourceKind.API)}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.API]}
            size="16px"
          />
          API
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex gap-x-2"
          onclick={() => {
            showAiConnectorDialog = true;
          }}
        >
          <Bot size="14px" class="stroke-icon-muted" />
          {m.status_label_ai_connector()}
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex gap-x-2"
          onclick={() =>
            createResourceAndNavigate(runtimeClient, ResourceKind.Theme)}
        >
          <svelte:component
            this={resourceIconMapping[ResourceKind.Theme]}
            size="16px"
          />
          {m.theme_label()}
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item class="flex gap-x-2" onclick={handleAddSkill}>
          <GraduationCap size="14px" class="stroke-icon-muted" />
          {m.add_asset_ai_skill()}
        </DropdownMenu.Item>
        <!-- Temporarily hide Report and Alert options -->
        <!-- <DropdownMenu.Item class="flex gap-x-2" onclick={() => createResourceAndNavigate(runtimeClient, ResourceKind.Report)}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Report]}
              className="text-fg-primary"
              size="16px"
            />
            {m.nav_tab_reports()}
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" onclick={() => createResourceAndNavigate(runtimeClient, ResourceKind.Alert)}>
            <svelte:component
              this={resourceIconMapping[ResourceKind.Alert]}
              className="text-fg-primary"
              size="16px"
            />
            {m.nav_tab_alerts()}
          </DropdownMenu.Item> -->
      </DropdownMenu.SubContent>
    </DropdownMenu.Sub>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<CreateExploreDialog bind:open={showExploreDialog} {metricsViews} />

<AddAiConnectorDialog bind:open={showAiConnectorDialog} />

<GenerateSampleData type="modal" bind:open={generateDataDialog} />

{#if $customComponents}
  <ChartTypesDialog bind:open={showChartTypesDialog} />
{/if}

<AddDataModal
  config={{
    medium: BehaviourEventMedium.Menu,
    space: MetricsEventSpace.LeftPanel,
    screen: screenName,
    targetResource: addDataTargetResource,
  }}
  bind:open={addDataModalOpen}
  connector={addDataConnector}
/>
