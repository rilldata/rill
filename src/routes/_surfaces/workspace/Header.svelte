<script lang="ts">
  import { getContext } from "svelte";
  import {
    ApplicationStore,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";

  import ModelIcon from "$lib/components/icons/Code.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import EditIcon from "$lib/components/icons/EditIcon.svelte";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
  import WorkspaceHeaderStatusSpinner from "./WorkspaceHeaderStatusSpinner.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  function formatModelName(str) {
    let output = str.trim().replaceAll(" ", "_").replace(/\.sql/, "");
    return output;
  }

  let currentModel: PersistentModelEntity;
  $: if ($store?.activeEntity && $persistentModelStore?.entities)
    currentModel = $persistentModelStore.entities.find(
      (q) => q.id === $store.activeEntity.id
    );

  let titleInputElement;
  let editingTitle = false;
  let titleInputValue;
  let tooltipActive;

  let clickOutsideListener;

  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  let titleInput = currentModel?.name;
  $: titleInput = currentModel?.name;

  function onKeydown(event) {
    if (editingTitle && event.key === "Enter") {
      titleInputElement.blur();
    }
  }

  let contextMenuOpen = false;
</script>

<svelte:window on:keydown={onKeydown} />

<header
  style:height="var(--header-height)"
  class="grid items-center content-stretch justify-between bg-gray-100 pl-6 pr-6"
  style:grid-template-columns="[title] auto [controls] auto"
>
  <div>
    {#if titleInput !== undefined && titleInput !== null}
      <h1
        style:font-size="16px"
        class="grid grid-flow-col justify-start items-center gap-x-1"
      >
        <ModelIcon />
        <Tooltip
          distance={8}
          bind:active={tooltipActive}
          suppress={editingTitle}
        >
          <input
            id="model-title-input"
            bind:this={titleInputElement}
            on:input={(evt) => {
              titleInputValue = evt.target.value;
              editingTitle = true;
            }}
            class="bg-gray-100 border border-transparent border-2 hover:border-gray-400 rounded pl-2 pr-2 cursor-pointer"
            class:font-bold={editingTitle === false}
            on:blur={() => {
              editingTitle = false;
            }}
            value={titleInput}
            size={Math.max(
              (editingTitle ? titleInputValue : titleInput)?.length || 0,
              5
            ) + 1}
            on:change={async (e) => {
              if (currentModel?.id) {
                const resp = await dataModelerService.dispatch(
                  "updateModelName",
                  [currentModel?.id, formatModelName(e.target.value)]
                );
                if (resp.status === ActionStatus.Failure) {
                  e.target.value = currentModel.name;
                }
              }
            }}
          />
          <TooltipContent slot="tooltip-content">
            <div class="flex items-center"><EditIcon size=".75em" />Edit</div>
          </TooltipContent>
        </Tooltip>
      </h1>
    {/if}
  </div>
  <WorkspaceHeaderStatusSpinner />
</header>
