<script lang="ts" context="module">
  export const editingIndex = writable<number | null>(null);
  export const editingType = writable<"measures" | "dimensions" | null>(null);
  export const editingField = writable<string | null>(null);

  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

  export class YAMLDimension {
    column: string | undefined;
    expression: string | undefined;
    name: string | undefined;
    label: string | undefined;
    description: string | undefined;
    unnest: boolean | undefined;

    constructor(item?: YAMLMap<string, string>) {
      this.column = item?.get("column");
      this.expression = item?.get("expression");
      this.name = item?.get("name");
      this.label = item?.get("label");
      this.description = item?.get("description");
      this.unnest = item?.get("unnest") as unknown as boolean;
    }
  }

  export class YAMLMeasure {
    expression: string | undefined;
    name: string | undefined;
    label: string | undefined;
    description: string | undefined;
    valid_percent_of_total: boolean | undefined;
    format_d3: string | undefined;
    format_preset: FormatPreset | undefined;

    constructor(item?: YAMLMap<string, string>) {
      this.expression = item?.get("expression");
      this.name = item?.get("name");
      this.label = item?.get("label");
      this.description = item?.get("description");
      this.valid_percent_of_total = item?.get(
        "valid_percent_of_total",
      ) as unknown as boolean;
      this.format_d3 = item?.get("format_d3");
      this.format_preset = item?.get(
        "format_preset",
      ) as unknown as FormatPreset;
    }
  }
</script>

<script lang="ts">
  import { createQueryServiceTableColumns } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { PlusIcon } from "lucide-svelte";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import MetricsTable from "../visual-metrics-editing/MetricsTable.svelte";
  import Sidebar from "../visual-metrics-editing/Sidebar.svelte";
  import { writable } from "svelte/store";
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

  type ItemType = "measures" | "dimensions";

  export let fileArtifact: FileArtifact;
  export let switchView: () => void;

  let searchValue = "";
  let confirmation: {
    action: "cancel" | "delete" | "switch";
    type?: ItemType;
    model?: string;
    index?: number;
  } | null = null;
  let collapsed = {
    measures: false,
    dimensions: false,
  };

  $: ({ instanceId } = $runtime);

  $: ({
    remoteContent,
    localContent,
    saveContent,
    updateLocalContent,
    saveLocalContent,
  } = fileArtifact);

  $: modelsQuery = useModels(instanceId);
  $: sourcesQuery = useSources(instanceId);

  $: modelNames =
    $modelsQuery.data
      ?.map((resource) => {
        return resource.meta?.name?.name;
      })
      .filter(isDefined) ?? [];
  $: sourceNames =
    $sourcesQuery.data
      ?.map((resource) => {
        return resource.meta?.name?.name;
      })
      .filter(isDefined) ?? [];

  $: timeDimension = (parsedDocument.get("timeseries") ?? "") as string;

  $: databaseSchema = (parsedDocument.get("database_schema") ?? "") as string;
  $: model = (parsedDocument.get("model") ??
    parsedDocument.get("table") ??
    "") as string;

  $: resourceQuery = useResource(instanceId, model, ResourceKind.Model);
  $: modelResource = $resourceQuery?.data?.model;
  $: connector = modelResource?.spec?.outputConnector;

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
    .map(({ name }) => name)
    .filter(isDefined);

  function isDefined<T>(x: T | undefined): x is T {
    return x !== undefined;
  }

  $: parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

  $: yamlMeasures = (
    (parsedDocument.get("measures") as YAMLSeq).items as Array<
      YAMLMap<string, string>
    >
  ).filter((i) => filter(i, searchValue));

  $: yamlDimensions = (
    (parsedDocument.get("dimensions") as YAMLSeq).items as Array<
      YAMLMap<string, string>
    >
  ).filter((i) => filter(i, searchValue));

  $: itemGroups = new Map<ItemType, YAMLMap<string, string>[]>([
    ["measures", yamlMeasures],
    ["dimensions", yamlDimensions],
  ]);

  $: smallestTimeGrain = (parsedDocument.get("smallest_time_grain") ??
    "") as string;

  function filter(item: YAMLMap<string, string>, searchValue: string) {
    return (
      item?.get("name")?.toLowerCase().includes(searchValue.toLowerCase()) ||
      item?.get("label")?.toLowerCase().includes(searchValue.toLowerCase()) ||
      item
        ?.get("expression")
        ?.toLowerCase()
        .includes(searchValue.toLowerCase()) ||
      item?.get("column")?.toLowerCase().includes(searchValue.toLowerCase())
    );
  }

  async function updateProperty(property: string, value: unknown) {
    parsedDocument.set(property, value);

    await saveContent(parsedDocument.toString());
  }

  async function reorderList(
    initIndex: number,
    newIndex: number,
    type: ItemType,
  ) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const clampedIndex = clamp(0, newIndex, items.length - 1);

    items.splice(clampedIndex, 0, items.splice(initIndex, 1)[0]);

    selected[type].delete(initIndex);
    selected[type].add(clampedIndex);

    selected = selected;

    parsedDocument.set(type, items);

    await saveContent(parsedDocument.toString());

    eventBus.emit("notification", { message: "Item moved", type: "success" });
  }

  function triggerDelete(index?: number, type?: ItemType) {
    confirmation = {
      action: "delete",
      index,
      type,
    };
  }

  async function deleteItems(items: Partial<typeof selected>) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    Object.entries(items).forEach(([type, indices]) => {
      const seq = parsedDocument.get(type) as YAMLSeq;
      const items = seq.items as Array<YAMLMap>;

      parsedDocument.set(
        type,
        items.filter((_, i) => !indices.has(i)),
      );

      indices.forEach((i) => {
        selected[type].delete(i);
      });
    });

    updateLocalContent(parsedDocument.toString(), true);

    selected = selected;

    await saveLocalContent();

    eventBus.emit("notification", { message: "Item deleted", type: "success" });
  }

  async function duplicateItem(item: number, type: ItemType) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const originalItem = items[item];
    const originalName = originalItem.get("name");
    const newItem = originalItem.clone() as YAMLMap;

    const itemNames = items.map((i) => i.get("name"));
    let count = 0;
    let newName = `${originalName}_copy`;
    newItem.set("name", newName);

    while (itemNames.includes(newName)) {
      count++;
      newName = `${originalName}_copy_${count}`;
      newItem.set("name", newName);
    }

    items.splice(item + 1, 0, newItem);

    parsedDocument.set(type, items);

    await saveContent(parsedDocument.toString());

    eventBus.emit("notification", {
      message: "Item duplicated",
      type: "success",
    });
  }

  $: item =
    ($editingType !== null &&
      $editingIndex !== null &&
      itemGroups.get($editingType)?.[$editingIndex]) ||
    undefined;

  $: editingClone =
    $editingIndex !== null
      ? $editingType === "measures"
        ? new YAMLMeasure(item)
        : new YAMLDimension(item)
      : undefined;

  let selected = {
    measures: new Set<number>(),
    dimensions: new Set<number>(),
  };

  $: totalSelected = selected.measures.size + selected.dimensions.size;
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex gap-x-4">
      {#key confirmation}
        <Input
          sameWidth
          full
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
        value={timeDimension}
        options={timeOptions}
        label="Time column"
        hint="Column from model that will be used as primary time dimension in dashboards"
        onInput={async (value) => {
          await updateProperty("timeseries", value);
        }}
      />

      <Input
        sameWidth
        full
        value={smallestTimeGrain}
        options={Object.entries(TIME_GRAIN).map(([_, { label }]) => label)}
        label="Smallest time grain"
        hint="The smallest time unit by which your charts and tables can be bucketed"
        onInput={async (value) => {
          await updateProperty("smallest_time_grain", value);
        }}
      />
    </div>

    <span class="h-[1px] w-full bg-gray-200" />

    <Input
      width="320px"
      textClass="text-sm"
      placeholder="Search"
      bind:value={searchValue}
      onInput={(value) => {
        searchValue = value;
      }}
    >
      <Search slot="icon" size="16px" color="#374151" />
    </Input>

    <div
      class="flex flex-col gap-y-2 h-fit w-full flex-shrink overflow-y-scroll"
    >
      {#each itemGroups as [type, items] (type)}
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
                editingType.set(type);
              }}
            >
              <PlusIcon size="16px" />
            </Button>
          </header>
          {#if !collapsed[type]}
            <MetricsTable
              selected={selected[type]}
              {reorderList}
              {type}
              onDuplicate={duplicateItem}
              {items}
              onDelete={triggerDelete}
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

    {#if totalSelected}
      <div
        class="bg-black rounded-[2px] z-20 shadow-md flex gap-x-0 h-8 text-gray-50 border border-gray-600 absolute bottom-10 left-1/2 -translate-x-1/2"
      >
        <div class="px-2 flex items-center">
          {totalSelected}
          {totalSelected > 1 ? "items" : "item"} selected:
        </div>
        <button
          on:click={async () => {
            triggerDelete();
          }}
          class="flex gap-x-2 text-inherit items-center px-2 border-l border-gray-600 hover:bg-gray-800 cursor-pointer"
        >
          <Trash size="16px" color="white" />
          Delete
        </button>

        <button
          on:click={() => {
            selected = {
              measures: new Set(),
              dimensions: new Set(),
            };
          }}
          class="flex gap-x-2 text-inherit items-center px-2 border-l border-gray-600 hover:bg-gray-800 cursor-pointer"
        >
          <Close size="14px" />
        </button>
      </div>
    {/if}
  </div>

  {#if $editingIndex !== null && $editingType !== null && editingClone}
    {#key editingClone}
      <Sidebar
        {item}
        {editingClone}
        {columns}
        onDelete={() => {
          if ($editingType) triggerDelete($editingIndex, $editingType);
        }}
        onCancel={(unsavedChanges) => {
          if (unsavedChanges) {
            confirmation = {
              action: "cancel",
              index: $editingIndex,
              type: $editingType,
            };
          } else {
            editingField.set(null);
            editingIndex.set(null);
            editingType.set(null);
          }
        }}
        index={$editingIndex}
        type={$editingType}
        field={$editingField}
        editing={$editingIndex !== -1}
        {fileArtifact}
        {switchView}
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
            <h2>Delete this {confirmation.type?.slice(0, -1)}?</h2>
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
            You will permanently remove this {confirmation.type?.slice(0, -1)} from
            all associated dashboards.
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
              } else if (confirmation?.action === "switch") {
                await updateProperty("model", confirmation.model);
              }
              confirmation = null;
              editingIndex.set(null);
              editingType.set(null);
            }}
          >
            {#if confirmation.action === "delete"}Yes, delete{:else if confirmation.action === "switch"}Switch
              model{:else}
              Close{/if}
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
