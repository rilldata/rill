<script lang="ts" context="module">
  export const editingItem: Writable<{
    index: number;
    type: "measures" | "dimensions";
    field?: string | null;
  } | null> = writable(null);

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
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { PlusIcon } from "lucide-svelte";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import MetricsTable from "../visual-metrics-editing/MetricsTable.svelte";
  import Sidebar from "../visual-metrics-editing/Sidebar.svelte";
  import { writable, Writable } from "svelte/store";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";

  export let fileArtifact: FileArtifact;
  export let switchView: () => void;

  let searchValue = "";
  let measuresCollapsed = false;
  let dimensionsCollapsed = false;
  let confirmation: {
    action: "cancel" | "delete";
    index: number;
    type: "measures" | "dimensions";
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: ({
    remoteContent,
    localContent,
    saveContent,
    updateLocalContent,
    saveLocalContent,
  } = fileArtifact);

  $: timeDimension = (parsedDocument.get("timeseries") ?? "") as string;

  $: databaseSchema = (parsedDocument.get("database_schema") ?? "") as string;
  $: table = (parsedDocument.get("model") ??
    parsedDocument.get("table") ??
    "") as string;

  $: resourceQuery = useResource(instanceId, table, ResourceKind.Model);
  $: modelResource = $resourceQuery?.data?.model;
  $: connector = modelResource?.spec?.outputConnector;

  $: columnsQuery = createQueryServiceTableColumns(instanceId, table, {
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
    .map(({ name }) => ({ value: name }));

  $: selected = timeOptions.find((option) => option.value === timeDimension);

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

  $: itemGroups = new Map([
    ["measures", yamlMeasures],
    ["dimensions", yamlDimensions],
  ]) as Map<"measures" | "dimensions", Array<YAMLMap<string, string>>>;

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

  async function handleColumnSelection(column: string) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    parsedDocument.set("timeseries", column);
    updateLocalContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }

  async function reorderList(
    initIndex: number,
    newIndex: number,
    type: "measures" | "dimensions",
  ) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const clampedIndex = clamp(0, newIndex, items.length - 1);

    items.splice(clampedIndex, 0, items.splice(initIndex, 1)[0]);

    parsedDocument.set(type, items);

    await saveContent(parsedDocument.toString());
  }

  function triggerDelete(index: number, type: "measures" | "dimensions") {
    confirmation = {
      action: "delete",
      index,
      type,
    };
  }

  async function deleteItem(index: number, type: "measures" | "dimensions") {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    items.splice(index, 1);

    parsedDocument.set(type, items);

    updateLocalContent(parsedDocument.toString(), true);

    await saveLocalContent();
  }

  async function duplicateItem(item: number, type: "measures" | "dimensions") {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const newItem = items[item].clone() as YAMLMap;

    if (type === "measures")
      newItem.set("name", `${newItem.get("name")}_copy_${items.length}`);

    items.splice(item + 1, 0, newItem);

    parsedDocument.set(type, items);

    await saveContent(parsedDocument.toString());
  }
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex flex-col gap-y-1">
      <span class="flex items-center gap-x-1">
        <p>Time column</p>
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            Time column description
          </TooltipContent>
        </Tooltip>
      </span>

      <Select.Root
        {selected}
        onSelectedChange={(newSelection) => {
          if (newSelection?.value) handleColumnSelection(newSelection.value);
        }}
        items={timeOptions}
      >
        <Select.Trigger class="w-[300px] rounded-[2px] shadow-none">
          <Select.Value placeholder={timeDimension} class="text-[12px]" />
        </Select.Trigger>
        <Select.Content>
          {#each timeOptions as { value } (value)}
            <Select.Item {value} class="text-[12px]">
              {value}
            </Select.Item>
          {/each}
        </Select.Content>
      </Select.Root>
    </div>

    <span class="h-[1px] w-full bg-gray-200" />

    <div class="flex gap-x-2">
      <form class="relative w-[320px] h-7">
        <div class="flex absolute inset-y-0 items-center pl-2 ui-copy-icon">
          <Search />
        </div>
        <input
          type="text"
          autocomplete="off"
          class="border outline-none rounded-[2px] block w-full pl-8 p-1"
          placeholder="Search"
          bind:value={searchValue}
        />
      </form>
    </div>

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
                if (type === "measures") measuresCollapsed = !measuresCollapsed;
                if (type === "dimensions")
                  dimensionsCollapsed = !dimensionsCollapsed;
              }}
            >
              <CaretDownIcon size="18px" className="!fill-gray-700" />
            </Button>
            <h1 class="capitalize font-medium">{type}</h1>
            <Button
              type="ghost"
              square
              gray
              noStroke
              on:click={() => {
                editingItem.set({
                  index: -1,
                  type,
                });
              }}
            >
              <PlusIcon size="16px" />
            </Button>
          </header>
          {#if (!measuresCollapsed && type === "measures") || (!dimensionsCollapsed && type === "dimensions")}
            <MetricsTable
              {reorderList}
              dimensions={type === "dimensions"}
              onDuplicate={duplicateItem}
              {items}
              onDelete={triggerDelete}
            />
          {/if}
        </div>
      {/each}
    </div>
  </div>

  {#if $editingItem}
    {#key $editingItem}
      <Sidebar
        {columns}
        item={itemGroups.get($editingItem.type)?.[$editingItem.index]}
        onDelete={() => {
          triggerDelete($editingItem.index, $editingItem.type);
        }}
        onCancel={(unsavedChanges) => {
          if (unsavedChanges) {
            confirmation = {
              action: "cancel",
              index: $editingItem.index,
              type: $editingItem.type,
            };
          } else {
            editingItem.set(null);
          }
        }}
        index={$editingItem.index}
        type={$editingItem.type}
        field={$editingItem.field}
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
            <h2>Delete this {confirmation.type.slice(0, -1)}?</h2>
          {:else}
            <h2>Cancel changes to {confirmation.type.slice(0, -1)}?</h2>
          {/if}</AlertDialog.Title
        >
        <AlertDialog.Description>
          {#if confirmation.action === "delete"}
            You haven't saved changes to this {confirmation.type.slice(0, -1)} yet,
            so closing this window will lose your work.
          {:else}
            You will permanently remove this {confirmation.type.slice(0, -1)} from
            all associated dashboards.
          {/if}
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer class="gap-y-2">
        <Button
          type="secondary"
          large
          on:click={() => {
            confirmation = null;
          }}
        >
          {#if confirmation.action === "delete"}Cancel{:else}Keep editing{/if}
        </Button>
        <Button
          large
          type="primary"
          on:click={async () => {
            if (confirmation?.action === "delete") {
              await deleteItem(confirmation.index, confirmation.type);
            }
            confirmation = null;
            editingItem.set(null);
          }}
        >
          {#if confirmation.action === "delete"}Yes, delete{:else}Close{/if}
        </Button>
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

  p {
    @apply font-medium text-sm;
  }

  h1 {
    @apply text-[16px] font-medium;
  }

  .main-area {
    @apply flex flex-col gap-y-4 size-full p-4 bg-background border;
    @apply flex-shrink overflow-hidden;
  }

  .section {
    @apply flex flex-col gap-y-2 justify-start w-full h-fit max-w-full;
  }

  h2 {
    @apply font-semibold text-lg;
  }
</style>
