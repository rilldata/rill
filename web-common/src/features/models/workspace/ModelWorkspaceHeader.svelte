<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconButton,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    useRuntimeServiceListCatalogEntries,
    useRuntimeServiceRenameFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import PanelCTA from "@rilldata/web-local/lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import { WorkspaceHeader } from "@rilldata/web-local/lib/components/workspace";
  import {
    isDuplicateName,
    renameFileArtifact,
    useAllNames,
  } from "@rilldata/web-local/lib/svelte-query/actions";
  import {
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "@rilldata/web-local/lib/util/entity-mappers";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let modelName: string;

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: allNamesQuery = useAllNames(runtimeInstanceId);
  const queryClient = useQueryClient();
  const renameModel = useRuntimeServiceRenameFileAndReconcile();

  const outputLayout = getContext("rill:app:output-layout");
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;
  $: modelHasError = !!modelError;

  let contextMenuOpen = false;

  $: availableDashboards = useRuntimeServiceListCatalogEntries(
    $runtimeStore.instanceId,
    { type: "OBJECT_TYPE_METRICS_VIEW" },
    {
      query: {
        select(data) {
          return data?.entries?.filter(
            (entry) => entry?.metricsView?.model === modelName
          );
        },
      },
    }
  );

  const onExport = async (exportExtension: "csv" | "parquet") => {
    // TODO: how do we handle errors ?
    window.open(
      `${RuntimeUrl}/v1/instances/${$runtimeStore.instanceId}/table/${modelName}/export/${exportExtension}`
    );
  };

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Model name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = modelName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, modelName, $allNamesQuery.data)) {
      notifications.send({
        message: `Name ${e.target.value} is already in use`,
      });
      e.target.value = modelName; // resets the input
      return;
    }

    try {
      const toName = e.target.value;
      const entityType = EntityType.Model;
      await renameFileArtifact(
        queryClient,
        runtimeInstanceId,
        modelName,
        toName,
        entityType,
        $renameModel
      );
      goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  $: titleInput = modelName;
</script>

<WorkspaceHeader
  let:width
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
  showStatus={false}
>
  <svelte:fragment slot="workspace-controls">
    <IconButton
      on:click={() => {
        outputLayout.update((state) => {
          state.visible = !state.visible;
          return state;
        });
      }}
      ><span class="text-gray-500"><HideBottomPane size="18px" /></span>
      <svelte:fragment slot="tooltip-content">
        <SlidingWords active={$outputLayout?.visible} reverse
          >results preview</SlidingWords
        >
      </svelte:fragment>
    </IconButton>
  </svelte:fragment>
  <svelte:fragment slot="cta">
    {@const collapse = width < 800}

    <PanelCTA side="right">
      <Tooltip
        alignment="middle"
        distance={16}
        location="left"
        suppress={contextMenuOpen}
      >
        <!-- attach floating element right here-->
        <WithTogglableFloatingElement
          alignment="end"
          bind:active={contextMenuOpen}
          distance={8}
          let:toggleFloatingElement
          location="bottom"
        >
          <Button
            disabled={modelHasError}
            on:click={toggleFloatingElement}
            type="secondary"
          >
            <IconSpaceFixer pullLeft pullRight={collapse}
              ><CaretDownIcon /></IconSpaceFixer
            >

            <ResponsiveButtonText {collapse}>Export</ResponsiveButtonText>
          </Button>
          <Menu
            dark
            on:click-outside={toggleFloatingElement}
            on:escape={toggleFloatingElement}
            slot="floating-element"
          >
            <MenuItem
              on:select={() => {
                toggleFloatingElement();
                onExport("parquet");
              }}
            >
              Export as Parquet
            </MenuItem>
            <MenuItem
              on:select={() => {
                toggleFloatingElement();
                onExport("csv");
              }}
            >
              Export as CSV
            </MenuItem>
          </Menu>
        </WithTogglableFloatingElement>
        <TooltipContent slot="tooltip-content">
          {#if modelHasError}Fix the errors in your model to export
          {:else}
            Export the modeled data as a file
          {/if}
        </TooltipContent>
      </Tooltip>

      {#if $availableDashboards?.data?.length === 0}
        <CreateDashboardButton
          collapse={width < 800}
          hasError={modelHasError}
          {modelName}
        />
      {:else if $availableDashboards?.data?.length === 1}
        <Button
          on:click={() => {
            goto(`/dashboard/${$availableDashboards.data[0].name}`);
          }}
        >
          <IconSpaceFixer pullLeft pullRight={collapse}>
            <Forward />
          </IconSpaceFixer>
          <ResponsiveButtonText {collapse}>
            Go to Dashboard
          </ResponsiveButtonText>
        </Button>
      {:else}
        <WithTogglableFloatingElement
          let:toggleFloatingElement
          distance={8}
          alignment="end"
        >
          <Button on:click={toggleFloatingElement}>
            <IconSpaceFixer pullLeft pullRight={collapse}>
              <Forward /></IconSpaceFixer
            >
            <ResponsiveButtonText {collapse}>
              Go to Dashboard
            </ResponsiveButtonText>
          </Button>
          <Menu
            dark
            slot="floating-element"
            on:escape={toggleFloatingElement}
            on:click-outside={toggleFloatingElement}
          >
            {#each $availableDashboards?.data as dashboard}
              <MenuItem
                on:select={() => {
                  goto(`/dashboard/${dashboard.name}`);
                  toggleFloatingElement();
                }}
              >
                {dashboard.name}
              </MenuItem>
            {/each}
          </Menu>
        </WithTogglableFloatingElement>
      {/if}
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>
