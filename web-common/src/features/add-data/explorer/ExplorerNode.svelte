<script lang="ts">
  import { builderActions, getAttrs } from "bits-ui";
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import {
    type ConnectorExplorerEntry,
    type ConnectorExplorerNode,
    ConnectorExplorerNodeType,
  } from "@rilldata/web-common/features/add-data/explorer/tree.ts";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { Database, Folder } from "lucide-svelte";
  import TableIcon from "@rilldata/web-common/components/icons/TableIcon.svelte";

  export let node: ConnectorExplorerNode;
  export let forceExpand: boolean;
  export let level: number;
  export let selectedEntry: ConnectorExplorerEntry | undefined;
  export let onSelect: (entry: ConnectorExplorerEntry) => void;

  let expanded = false;
  function forceExpandNode() {
    expanded = true;
  }
  $: if (forceExpand) forceExpandNode();

  $: isLeafType = node.type === ConnectorExplorerNodeType.Table;

  const leftPad = level * 20;
  // pl-[20px] pl-[40px] pl-[60px] pl-[80px]
  const leftPadClass = `pl-[${leftPad}px]`;

  $: selected =
    selectedEntry &&
    node.entry.connector === selectedEntry.connector &&
    node.entry.database === selectedEntry.database &&
    node.entry.databaseSchema === selectedEntry.databaseSchema &&
    node.entry.table === selectedEntry.table;
</script>

<li>
  {#if node.error}
    <span class="message {leftPadClass}">Error: {node.error}</span>
  {:else if node.loading}
    <span class="message {leftPadClass}">Loading...</span>
  {:else if isLeafType}
    <button
      type="button"
      class="entry {leftPadClass}"
      class:bg-gray-100={selected}
      on:click={() => onSelect(node.entry)}
      aria-label="Node: {node.name}, level {level}"
    >
      <TableIcon size="14px" className="shrink-0 text-fg-secondary" />
      <span>{node.name}</span>
    </button>
  {:else}
    <Collapsible.Root bind:open={expanded}>
      <Collapsible.Trigger asChild let:builder>
        <button
          type="button"
          class="entry {leftPadClass}"
          class:bg-gray-100={selected}
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
          aria-label="Node: {node.name}, level {level}"
        >
          <CaretDownIcon
            className="transform transition-transform text-fg-secondary {expanded
              ? 'rotate-0'
              : '-rotate-90'}"
            size="14px"
          />
          {#if node.type === ConnectorExplorerNodeType.Database}
            <Database size="14px" class="shrink-0 text-fg-secondary" />
          {:else if node.type === ConnectorExplorerNodeType.Schema}
            <Folder size="14px" class="shrink-0 text-fg-secondary" />
          {/if}
          <span class="truncate text-fg-primary">
            {node.name}
          </span>
        </button>
      </Collapsible.Trigger>
      <Collapsible.Content>
        <ol>
          {#if node.children?.length}
            {#each node.children as child (child.name)}
              <svelte:self
                node={child}
                {forceExpand}
                level={level + 1}
                {selectedEntry}
                {onSelect}
              />
            {/each}
          {:else}
            <span class="message">No tables found</span>
          {/if}
        </ol>
      </Collapsible.Content>
    </Collapsible.Root>
  {/if}
</li>

<style lang="postcss">
  .message {
    @apply text-fg-secondary pl-6 py-1;
  }

  .entry {
    @apply flex flex-row items-center gap-1.5 w-full py-1 cursor-pointer;
  }
  .entry:hover {
    @apply bg-surface-hover;
  }
</style>
