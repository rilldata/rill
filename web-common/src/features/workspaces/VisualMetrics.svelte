<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createQueryServiceTableColumns } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { FileArtifact } from "../entity-management/file-artifact";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { parseDocument } from "yaml";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { Plus, SearchIcon } from "lucide-svelte";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  //   export let selectedColumn: string;
  export let fileArtifact: FileArtifact;

  $: resource = fileArtifact.getResource(queryClient, $runtime.instanceId);

  $: ({ remoteContent, localContent, updateLocalContent, saveLocalContent } =
    fileArtifact);

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
  $: ({ data: columnsResponse, error, isError } = $columnsQuery);

  $: columns = columnsResponse?.profileColumns ?? [];

  async function handleColumnSelection(column: string) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    parsedDocument.set("timeseries", column);
    updateLocalContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }

  //   async function handlePreviewUpdate(
  //     e: CustomEvent<{
  //       index: number;
  //       position: Vector;
  //       dimensions: Vector;
  //     }>,
  //   ) {
  //     const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
  //     const items = parsedDocument.get("items") as any;

  //     const node = items.get(e.detail.index);

  //     node.set("width", e.detail.dimensions[0]);
  //     node.set("height", e.detail.dimensions[1]);
  //     node.set("x", e.detail.position[0]);
  //     node.set("y", e.detail.position[1]);

  //     updateLocalContent(parsedDocument.toString(), true);

  //     if ($autoSave) await updateChartFile();
  //   }
</script>

<div class="flex flex-col gap-y-4 bg-gray-50">
  <div class="flex flex-col gap-y-1">
    <span class="flex items-center gap-x-1">
      <p>Time column</p>
      <InfoCircle size="12px" color="#6B7280" />
    </span>

    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        <!-- <IconButton>
        <ThreeDot size="16px" />
      </IconButton> -->
        <button
          class="bg-white w-[300px] h-8 border justify-between flex items-center px-3 hover:bg-gray-50 rounded-[2px]"
        >
          {timeDimension}
          <CaretDownIcon />
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="start"
        class="max-h-96 w-[300px] overflow-scroll"
      >
        {#each columns as { name, type } (name)}
          {#if type === "TIMESTAMP" && name}
            <DropdownMenu.Item on:click={() => handleColumnSelection(name)}>
              {name}
            </DropdownMenu.Item>
          {/if}
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>

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
        on:input
      />
    </form>

    <div class="h-7 flex items-center border border-primary-50">
      <Button size="medium" type="secondary"><Plus size="14px" />Add</Button>
      <Button type="secondary" square icon><CaretDownIcon /></Button>
    </div>
  </div>
</div>

<style lang="postcss">
  p {
    @apply font-medium text-sm;
  }
</style>
