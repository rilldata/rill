<script lang="ts">
  import {
    createQueryServiceTableColumns,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { PlusIcon } from "lucide-svelte";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import MetricsTable from "../visual-metrics-editing/MetricsTable.svelte";
  import Sidebar from "../visual-metrics-editing/Sidebar.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { useModels } from "../models/selectors";
  import { useSources } from "../sources/selectors";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { tick } from "svelte";
  import type { ItemType, Confirmation } from "../visual-metrics-editing/lib";
  import {
    YAMLMeasure,
    YAMLDimension,
    editingIndex,
    editingItem,
    types,
  } from "../visual-metrics-editing/lib";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let fileArtifact: FileArtifact;
  export let errors: LineStatus[];
  export let switchView: () => void;

  let searchValue = "";
  let unsavedChanges = false;
  let confirmation: Confirmation | null = null;
  let collapsed = {
    measures: false,
    dimensions: false,
  };
  let selected = {
    measures: new Set<number>(),
    dimensions: new Set<number>(),
  };

  $: ({ instanceId } = $runtime);

  $: totalSelected = selected.measures.size + selected.dimensions.size;

  $: ({ remoteContent, localContent, saveContent, getResource } = fileArtifact);

  // YAML Parsing
  $: parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

  $: raw = {
    measures: parsedDocument.get("measures"),
    dimensions: parsedDocument.get("dimensions"),
  };

  $: rawSmallestTimeGrain = parsedDocument.get("smallest_time_grain");
  $: rawTimeDimension = parsedDocument.get("timeseries");
  $: rawDatabaseSchema = parsedDocument.get("database_schema");
  $: rawModel = parsedDocument.get("model");
  $: rawTable = parsedDocument.get("table");

  $: timeDimension = stringGuard(rawTimeDimension);
  $: databaseSchema = stringGuard(rawDatabaseSchema);
  $: model = stringGuard(rawModel) || stringGuard(rawTable);

  $: itemGroups = {
    measures:
      raw.measures instanceof YAMLSeq
        ? raw.measures.items.map((item) => new YAMLMeasure(item))
        : [],
    dimensions:
      raw.dimensions instanceof YAMLSeq
        ? raw.dimensions.items.map(
            (item, i) => new YAMLDimension(item, dimensions[i]),
          )
        : [],
  };

  $: smallestTimeGrain =
    rawSmallestTimeGrain && typeof rawSmallestTimeGrain === "string"
      ? rawSmallestTimeGrain
      : undefined;

  // Queries
  $: modelsQuery = useModels(instanceId);
  $: sourcesQuery = useSources(instanceId);
  $: metricsViewQuery = getResource(queryClient, instanceId);
  $: resourceQuery = useResource(instanceId, model, ResourceKind.Model);

  $: modelNames = $modelsQuery?.data?.map(resourceToOption) ?? [];
  $: sourceNames = $sourcesQuery?.data?.map(resourceToOption) ?? [];
  $: dimensions = $metricsViewQuery?.data?.metricsView?.spec?.dimensions ?? [];
  $: connector = $resourceQuery?.data?.model?.spec?.outputConnector;

  $: columnsQuery = createQueryServiceTableColumns(instanceId, model, {
    connector,
    database: "", // models use the default database
    databaseSchema,
  });

  $: ({ data: columnsResponse } = $columnsQuery);

  $: columns = columnsResponse?.profileColumns ?? [
    { name: timeDimension, type: "TIMESTAMP" },
  ];

  $: timeOptions = columns
    .filter(({ type }) => type === "TIMESTAMP")
    .map(({ name }) => ({ value: name ?? "", label: name ?? "" }));

  /** display the main error (the first in this array) at the bottom */
  $: mainError = errors?.at(0);

  function stringGuard(value: unknown | undefined): string {
    return value && typeof value === "string" ? value : "";
  }

  async function setEditing(index: number, type: ItemType, field?: string) {
    if (unsavedChanges) {
      confirmation = {
        action: "cancel",
        index,
        type,
        field,
      };
      return;
    }

    const item = itemGroups[type][index];

    if (item) {
      editingItem.set({ item, type });
      editingIndex.set(index);
    }

    if (field) {
      await tick();
      document.getElementById(`vme-${field}`)?.focus();
    }
  }

  function resetEditing() {
    editingItem.set(null);
    editingIndex.set(null);
    unsavedChanges = false;
  }

  async function updateProperty(
    property: string,
    value: unknown,
    removeProperty?: string,
  ) {
    parsedDocument.set(property, value);

    if (removeProperty) {
      parsedDocument.delete(removeProperty);
    }

    await saveContent(parsedDocument.toString());
  }

  function resourceToOption(resource: V1Resource) {
    const value = resource.meta?.name?.name ?? "";
    return {
      value,
      label: value,
    };
  }

  async function reorderList(
    initIndexes: number[],
    newIndex: number,
    type: ItemType,
  ) {
    initIndexes.sort((a, b) => a - b);
    const editingItemIndex = initIndexes.indexOf($editingIndex ?? -1);

    const sequence = raw[type];

    if (!(sequence instanceof YAMLSeq)) {
      return;
    }

    let items = sequence.items as Array<YAMLMap | null>;

    const clampedIndex = clamp(0, newIndex, items.length);

    const movedItems: Array<YAMLMap | null> = [];

    initIndexes.forEach((index) => {
      movedItems.push(items[index]);
      items[index] = null;
    });

    items.splice(clampedIndex, 0, ...movedItems);

    const countBeforeClamped = initIndexes.filter(
      (i) => i < clampedIndex,
    ).length;

    const newIndexes = initIndexes.map((_, dragPosition) => {
      return clampedIndex + dragPosition - countBeforeClamped;
    });

    if (editingItemIndex !== -1) {
      editingIndex.set(newIndexes[editingItemIndex]);
    }

    if (selected[type].size) {
      selected[type] = new Set(newIndexes);
    }

    await updateProperty(
      type,
      items.filter((i) => i !== null),
    );

    eventBus.emit("notification", { message: "Item moved", type: "success" });
  }

  function triggerDelete(index?: number, type?: ItemType) {
    if (totalSelected) {
      confirmation = {
        action: "delete",
      };
    } else {
      confirmation = {
        action: "delete",
        index,
        type,
      };
    }
  }

  async function deleteItems(items: Partial<typeof selected>) {
    let deletedEditingItem = false;

    Object.entries(items).forEach(([type, indices]) => {
      const sequence = raw[type];

      if (!(sequence instanceof YAMLSeq)) {
        return;
      }
      const items = sequence.items as Array<YAMLMap>;

      if ($editingIndex !== null) {
        deletedEditingItem =
          indices.has($editingIndex) && type === $editingItem?.type;
      }

      parsedDocument.set(
        type,
        items.filter((_, i) => !indices.has(i)),
      );

      indices.forEach((i) => {
        selected[type].delete(i);
      });
    });

    selected = selected;

    await saveContent(parsedDocument.toString());

    if (deletedEditingItem) {
      resetEditing();
    }

    eventBus.emit("notification", { message: "Item deleted", type: "success" });
  }

  async function duplicateItem(index: number, type: ItemType) {
    const sequence = raw[type];

    if (!(sequence instanceof YAMLSeq)) {
      return;
    }

    const items = sequence.items as Array<YAMLMap>;

    const originalItem = items[index];
    const name = stringGuard(originalItem.get("name"));
    const split = name.split("_copy");
    const potentialNumber = split[1] ? split[1].split("_")[1] : undefined;
    const number: number | undefined = isNaN(Number(potentialNumber))
      ? undefined
      : Number(potentialNumber);

    const canUseOriginalName =
      split.length <= 2 && (!split[1] || number !== undefined);

    const originalName = canUseOriginalName ? split[0] : name;
    const newItem = originalItem.clone() as YAMLMap;

    const itemNames = items.map((i) => stringGuard(i.get("name")));
    let count = number ?? 0;
    let newName = `${originalName}_copy`;
    newItem.set("name", newName);

    while (itemNames.includes(newName)) {
      count++;
      newName = `${originalName}_copy_${count}`;
      newItem.set("name", newName);
    }

    items.splice(index + 1, 0, newItem);

    await updateProperty(type, items);

    eventBus.emit("notification", {
      message: "Item duplicated",
      type: "success",
    });
  }

  $: dimensionNamesAndLabels = itemGroups.dimensions.reduce(
    (acc, { name, label }) => {
      acc.name = Math.max(acc.name, name.length);
      acc.label = Math.max(acc.label, label.length);
      return acc;
    },
    { name: 0, label: 0 },
  );

  $: measureNamesAndLabels = itemGroups.measures.reduce(
    (acc, { name, label }) => {
      acc.name = Math.max(acc.name, name.length);
      acc.label = Math.max(acc.label, label.length);
      return acc;
    },
    { name: 0, label: 0 },
  );

  $: longestName = Math.max(
    dimensionNamesAndLabels.name,
    measureNamesAndLabels.name,
  );
  $: longestLabel = Math.max(
    dimensionNamesAndLabels.label,
    measureNamesAndLabels.label,
  );
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex gap-x-4">
      {#key confirmation}
        <Input
          sameWidth
          full
          truncate
          value={model}
          options={[...modelNames, ...sourceNames]}
          label="Model or source referenced"
          onChange={(newModelName) => {
            confirmation = {
              action: "switch",
              model: newModelName,
            };
          }}
        />
      {/key}

      <Input
        sameWidth
        full
        truncate
        value={timeDimension}
        options={timeOptions}
        label="Time column"
        hint="Column from model that will be used as primary time dimension in dashboards"
        onChange={async (value) => {
          await updateProperty("timeseries", value);
        }}
      />

      <Input
        sameWidth
        full
        truncate
        value={smallestTimeGrain}
        options={Object.entries(TIME_GRAIN).map(([_, { label }]) => ({
          value: label,
          label,
        }))}
        label="Smallest time grain"
        hint="The smallest time unit by which your charts and tables can be bucketed"
        onChange={async (value) => {
          await updateProperty("smallest_time_grain", value);
        }}
      />
    </div>

    <span class="h-[1px] w-full bg-gray-200" />

    <div class="grid grid-cols-3 gap-4 relative">
      <div class="col-span-3 sm:col-span-2 lg:col-span-1">
        <Input
          full
          textClass="text-sm"
          placeholder="Search"
          bind:value={searchValue}
          onInput={(value) => {
            searchValue = value;
          }}
        >
          <Search slot="icon" size="16px" color="#374151" />
        </Input>
      </div>

      {#if totalSelected}
        <div
          class="bg-white rounded-[2px] z-20 shadow-md flex gap-x-0 h-8 text-gray-700 border border-slate-100 absolute right-0"
        >
          <div class="px-2 flex items-center">
            {totalSelected}
            {totalSelected > 1 ? "items" : "item"} selected:
          </div>
          <button
            on:click={() => {
              triggerDelete();
            }}
            class="flex gap-x-2 text-inherit items-center px-2 border-l border-slate-100 hover:bg-gray-50 cursor-pointer"
          >
            <Trash size="16px" />
            Delete
          </button>

          <button
            on:click={() => {
              selected = {
                measures: new Set(),
                dimensions: new Set(),
              };
            }}
            class="flex gap-x-2 text-inherit items-center px-2 border-l border-slate-100 hover:bg-gray-50 cursor-pointer"
          >
            <Close size="14px" />
          </button>
        </div>
      {/if}
    </div>

    <div
      class="flex flex-col gap-y-4 h-fit w-full flex-shrink overflow-y-scroll"
    >
      {#each types as type (type)}
        {@const items = itemGroups[type]}
        <div class="section">
          <header class="flex gap-x-1 items-center">
            <Button
              type="ghost"
              square
              gray
              noStroke
              on:click={() => {
                collapsed[type] = !collapsed[type];
              }}
            >
              <span
                class="transition-transform"
                class:-rotate-90={collapsed[type]}
              >
                <CaretDownIcon size="16px" className="!fill-gray-700" />
              </span>
            </Button>
            <h1 class="capitalize font-medium">{type}</h1>
            <Button
              type="ghost"
              square
              gray
              noStroke
              on:click={() => {
                editingIndex.set(-1);
                editingItem.set({
                  type,
                  item:
                    type === "measures"
                      ? new YAMLMeasure()
                      : new YAMLDimension(),
                });
              }}
            >
              <PlusIcon size="16px" />
            </Button>
          </header>
          {#if !collapsed[type]}
            <MetricsTable
              selected={selected[type]}
              {type}
              {items}
              {searchValue}
              longest={{ name: longestName, label: longestLabel }}
              editingIndex={$editingIndex !== null &&
              $editingItem?.type === type
                ? $editingIndex
                : null}
              {reorderList}
              onDuplicate={duplicateItem}
              onDelete={triggerDelete}
              onEdit={setEditing}
              onCheckedChange={(checked, index) => {
                if (index === undefined) {
                  if (checked) {
                    selected[type] = new Set(
                      Array.from({ length: items.length }, (_, i) => i),
                    );
                  } else {
                    selected[type] = new Set();
                  }
                } else {
                  selected[type][checked ? "add" : "delete"](index);
                  selected[type] = selected[type];
                }
              }}
            />
          {/if}
        </div>
      {/each}
    </div>

    {#if mainError}
      <div
        role="status"
        transition:slide={{ duration: LIST_SLIDE_DURATION }}
        class=" flex items-center gap-x-2 ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5 max-h-40 overflow-auto"
      >
        <CancelCircle />
        {mainError.message}
      </div>
    {/if}
  </div>

  {#if $editingItem && $editingIndex !== null}
    {@const { item, type } = $editingItem}
    {@const index = $editingIndex}
    {#key $editingItem}
      <Sidebar
        {item}
        {columns}
        onDelete={() => {
          triggerDelete(index, type);
        }}
        onCancel={(unsavedChanges) => {
          if (unsavedChanges) {
            confirmation = {
              action: "cancel",
              index,
              type,
            };
          } else {
            resetEditing();
          }
        }}
        {index}
        {type}
        {resetEditing}
        editing={index !== -1}
        {fileArtifact}
        {switchView}
        bind:unsavedChanges
      />
    {/key}
  {/if}
</div>

{#if confirmation}
  <AlertDialog.Root open>
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>
          {#if confirmation.action === "delete"}
            <h2>
              Delete {confirmation.index === undefined
                ? "selected items"
                : "this " + confirmation.type?.slice(0, -1)}?
            </h2>
          {:else if confirmation.action === "cancel"}
            <h2>Cancel changes to {confirmation.type?.slice(0, -1)}?</h2>
          {:else if confirmation.action === "switch"}
            <h2>Switch reference model?</h2>
          {/if}
        </AlertDialog.Title>
        <AlertDialog.Description>
          {#if confirmation.action === "cancel"}
            You haven't saved changes to this {confirmation.type?.slice(0, -1)} yet,
            so closing this window will lose your work.
          {:else if confirmation.action === "delete"}
            You will permanently remove {confirmation.index === undefined
              ? "the selected items"
              : "this " + confirmation.type?.slice(0, -1)} from all associated dashboards.
          {:else if confirmation.action === "switch"}
            Switching to a different model may break your measures and
            dimensions unless the new model has similar data.
          {/if}
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer class="gap-y-2">
        <AlertDialog.Cancel asChild let:builder>
          <Button
            builders={[builder]}
            type="secondary"
            large
            gray={confirmation.action === "delete"}
            on:click={() => {
              confirmation = null;
            }}
          >
            {#if confirmation.action === "cancel"}Keep editing{:else}Cancel{/if}
          </Button>
        </AlertDialog.Cancel>

        <AlertDialog.Action asChild let:builder>
          <Button
            large
            builders={[builder]}
            status={confirmation.action === "delete" ? "error" : "info"}
            type="primary"
            on:click={async () => {
              if (confirmation?.action === "delete") {
                await deleteItems(
                  confirmation?.index !== undefined && confirmation.type
                    ? {
                        [confirmation.type]: new Set([confirmation.index]),
                      }
                    : selected,
                );

                resetEditing();
              } else if (confirmation?.action === "switch") {
                await updateProperty("model", confirmation.model, "table");
                resetEditing();
              } else if (confirmation?.action === "cancel") {
                if (
                  confirmation?.field &&
                  confirmation?.index !== undefined &&
                  confirmation?.type
                ) {
                  unsavedChanges = false;
                  await setEditing(
                    confirmation.index,
                    confirmation.type,
                    confirmation.field,
                  );
                } else {
                  resetEditing();
                }
              }
              confirmation = null;
            }}
          >
            {#if confirmation.action === "delete"}
              Yes, delete
            {:else if confirmation.action === "switch"}
              Switch model
            {:else if confirmation.action === "cancel" && confirmation.field}
              Switch items
            {:else}
              Close
            {/if}
          </Button>
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

<style lang="postcss">
  .wrapper {
    @apply size-full max-w-full max-h-full flex-none;
    @apply overflow-hidden;
    @apply flex gap-x-3 p-4;
  }

  h1 {
    @apply text-[16px] font-medium;
  }

  .main-area {
    @apply flex flex-col gap-y-4 size-full p-4 bg-background border;
    @apply flex-shrink overflow-hidden rounded-[2px] relative;
  }

  .section {
    @apply flex flex-col gap-y-2 justify-start w-full h-fit max-w-full;
  }

  h2 {
    @apply font-semibold text-lg;
  }
</style>
