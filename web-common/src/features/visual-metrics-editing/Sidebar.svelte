<script lang="ts" context="module">
  import { editingItem } from "../workspaces/VisualMetrics.svelte";

  const properties = {
    measures: [
      { label: "Definition", key: "expression" },
      { label: "Name", key: "name", hint: "Name hint" },
      { label: "Label", key: "label", hint: "Label hint", optional: true },
      {
        label: "Format",
        key: "formatPreset",
        yamlKey: "format_preset",
        modes: [
          {
            label: "Simple",
            type: "select",
            options: Object.values(FormatPreset),
          },
          {
            label: "d3-format",
            type: "text",
          },
        ],
      },
      { label: "Description", key: "description", optional: true },
    ],
    dimensions: [
      { label: "Definition", key: "expression" },
      { label: "Name", key: "name", hint: "Name hint" },
      { label: "Label", key: "label", hint: "Label hint", optional: true },
      { label: "Description", key: "description", optional: true },
    ],
  };
</script>

<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { FileArtifact } from "../entity-management/file-artifact";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

  export let editing: MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2;
  export let onDelete: () => void;
  export let fileArtifact: FileArtifact;
  export let index: number;
  export let type: "measures" | "dimensions";

  export let switchView: () => void;

  $: ({ remoteContent, localContent, saveContent } = fileArtifact);

  $: adding = index === -1;

  async function saveChanges() {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const items = (parsedDocument.get(type) as YAMLSeq).items as Array<YAMLMap>;

    if (adding) {
      const newItem = new YAMLMap();
      properties[type].forEach(({ key, yamlKey }) => {
        if (editing[key]) newItem.set(yamlKey || key, editing[key]);
      });
      items.push(newItem);
    } else {
      const item = items[index];

      properties[type].forEach(({ key, yamlKey }) => {
        item.set(yamlKey || key, editing[key]);
      });
    }

    await saveContent(parsedDocument.toString());
  }
</script>

<div class="h-full w-[320px] bg-background flex-none p-6 flex flex-col border">
  <h1>{adding ? "Add" : "Edit"} {type.slice(0, -1)}</h1>

  <div
    class="flex flex-col gap-y-3 w-full h-fit overflow-y-auto overflow-x-visible"
  >
    {#each properties[type] as { label, key, hint, optional, modes } (key)}
      <Input bind:value={editing[key]} {label} {hint} {optional} {modes} />
    {/each}

    <h2>Referenced by</h2>

    <span />

    <h2>Preview</h2>
  </div>

  <div class="flex flex-col gap-y-3 mt-auto">
    <p>
      For more options,
      <button on:click={switchView} class="text-primary-600 font-medium">
        edit in YAML
      </button>
    </p>
    <div class="flex justify-{adding ? 'end' : 'between'}">
      {#if !adding}
        <Button type="text" on:click={onDelete}>Delete</Button>
      {/if}
      <div class="flex gap-x-2 self-end">
        <Button
          type="secondary"
          on:click={() => {
            editingItem.set(null);
          }}
        >
          Cancel
        </Button>
        <Button type="primary" on:click={saveChanges}
          >{adding ? "Add " + type.slice(0, -1) : "Save changes"}</Button
        >
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  h1 {
    @apply text-lg font-medium mb-4;
  }

  h2 {
    @apply text-sm font-medium;
  }
</style>
