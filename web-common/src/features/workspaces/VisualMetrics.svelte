<script lang="ts" context="module">
  export const editingItem: Writable<{
    index: number;
    data: MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2;
    type: "measures" | "dimensions";
  } | null> = writable(null);
</script>

<script lang="ts">
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createQueryServiceTableColumns,
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { Plus } from "lucide-svelte";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import MetricsTable from "../visual-metrics-editing/MetricsTable.svelte";
  import Sidebar from "../visual-metrics-editing/Sidebar.svelte";
  import { writable, Writable } from "svelte/store";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import * as Select from "@rilldata/web-common/components/select";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  const options: ["measures", "dimensions"] = ["measures", "dimensions"];

  export let fileArtifact: FileArtifact;
  export let switchView: () => void;

  let searchValue = "";

  $: resource = fileArtifact.getResource(queryClient, $runtime.instanceId);

  $: ({
    remoteContent,
    localContent,
    saveContent,
    updateLocalContent,
    saveLocalContent,
  } = fileArtifact);

  $: ({ data } = $resource);

  $: timeDimension = data?.metricsView?.state?.validSpec?.timeDimension;

  $: connector = data?.metricsView?.state?.validSpec?.connector ?? "";
  $: database = data?.metricsView?.state?.validSpec?.database ?? "";
  $: databaseSchema = data?.metricsView?.state?.validSpec?.databaseSchema ?? "";
  $: table = data?.metricsView?.state?.validSpec?.table ?? "";

  $: columnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    table,
    {
      connector,
      database,
      databaseSchema,
    },
  );
  $: ({ data: columnsResponse } = $columnsQuery);

  $: columns = columnsResponse?.profileColumns ?? [];

  $: measures = data?.metricsView?.spec?.measures ?? [];
  $: dimensions = data?.metricsView?.spec?.dimensions ?? [];

  $: filteredMeasures = measures.filter(
    (measure) =>
      measure?.name?.toLowerCase().includes(searchValue.toLowerCase()) ||
      measure?.label?.toLowerCase().includes(searchValue.toLowerCase()),
  );

  $: filteredDimensions = dimensions.filter(
    (dimension) =>
      dimension?.name?.toLowerCase().includes(searchValue.toLowerCase()) ||
      dimension?.label?.toLowerCase().includes(searchValue.toLowerCase()) ||
      dimension?.expression?.toLowerCase().includes(searchValue.toLowerCase()),
  );

  async function handleColumnSelection(column: string) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    parsedDocument.set("timeseries", column);
    updateLocalContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }

  async function reorderList(initIndex: number, newIndex: number) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get("measures") as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const clampedIndex = clamp(0, newIndex, items.length - 1);

    items.splice(clampedIndex, 0, items.splice(initIndex, 1)[0]);

    parsedDocument.set("measures", items);

    await saveContent(parsedDocument.toString());
  }

  async function deleteItem(index: number) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get("measures") as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    items.splice(index, 1);

    parsedDocument.set("measures", items);

    updateLocalContent(parsedDocument.toString(), true);

    await saveLocalContent();
  }

  async function duplicateItem(item: number, type: "measures" | "dimensions") {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const measures = parsedDocument.get(type) as YAMLSeq;

    const items = measures.items as Array<YAMLMap>;

    const newItem = items[item].clone() as YAMLMap;

    if (type === "measures")
      newItem.set("name", `${newItem.get("name")} copy ${items.length}`);

    items.push(newItem);

    parsedDocument.set(type, items);

    await saveContent(parsedDocument.toString());
  }

  $: timeOptions = columns
    .filter(({ type }) => type === "TIMESTAMP")
    .map(({ name }) => ({ value: name }));

  $: selected = timeOptions.find((option) => option.value === timeDimension);
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

      <div
        class="h-7 flex items-center border border-primary-300 text-primary-600 rounded-[2px]"
      >
        <button
          class="flex gap-x-2 w-[72px] h-full items-center pl-3 text-primary-600 hover:bg-primary-100"
          on:click={() => {
            editingItem.set({
              index: -1,
              data: {},
              type: "measures",
            });
          }}
        >
          <Plus size="14px" />Add
        </button>

        <DropdownMenu.Root>
          <DropdownMenu.Trigger asChild let:builder>
            <button
              use:builder.action
              {...builder}
              class="aspect-square h-full grid place-content-center border-l border-primary-300 hover:bg-primary-100 text-primary-600"
            >
              <CaretDownIcon size="14px" />
            </button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content class="w-[100px] max-w-[100px]" align="end">
            {#each options as option (option)}
              <DropdownMenu.Item
                class="text-[12px] capitalize"
                on:click={() => {
                  editingItem.set({
                    index: -1,
                    data: {},
                    type: option,
                  });
                }}
              >
                {option.slice(0, -1)}
              </DropdownMenu.Item>
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>

    <div class="flex flex-col gap-y-2 h-full w-full flex-shrink overflow-auto">
      {#each options as option (option)}
        <div class="section">
          <h1 class="capitalize">{option}</h1>
          <MetricsTable
            {reorderList}
            dimensions={option === "dimensions"}
            onDuplicate={async (item) => {
              await duplicateItem(item, option);
            }}
            items={option === "dimensions"
              ? filteredDimensions
              : filteredMeasures}
            onDelete={deleteItem}
          />
          <!-- <Button type="text" fit>Show all {option}</Button> -->
        </div>
      {/each}
    </div>
  </div>

  {#if $editingItem}
    <Sidebar
      editing={$editingItem.data}
      onDelete={async () => {
        await deleteItem($editingItem.index);
        editingItem.set(null);
      }}
      index={$editingItem.index}
      type={$editingItem.type}
      {fileArtifact}
      {switchView}
    />
  {/if}
</div>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") editingItem.set(null);
  }}
/>

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
    @apply flex flex-col gap-y-2 justify-start size-full max-w-full;
  }
</style>
