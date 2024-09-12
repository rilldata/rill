<script lang="ts">
  import {
    editingItem,
    YAMLDimension,
    YAMLMeasure,
  } from "../workspaces/VisualMetrics.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { V1ProfileColumn } from "@rilldata/web-common/runtime-client";
  import { FileArtifact } from "../entity-management/file-artifact";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";

  export let item: YAMLMap<string, string> | undefined;
  export let onDelete: () => void;
  export let onCancel: (unsavedChanges: boolean) => void;
  export let fileArtifact: FileArtifact;
  export let index: number;
  export let type: "measures" | "dimensions";
  export let switchView: () => void;
  export let columns: V1ProfileColumn[];

  let editingClone =
    type === "measures" ? new YAMLMeasure(item) : new YAMLDimension(item);

  let columnNames = columns.map((column) => column.name).filter(isDefined);

  let properties: Record<
    string,
    Array<{
      label: string;
      selected: number;
      optional?: true;
      fields: Array<{
        key: string;
        hint?: string;
        label: string;

        options?: string[];
        placeholder?: string;
        boolean?: true;
      }>;
    }>
  > = {
    measures: [
      {
        label: "SQL Expression",

        fields: [
          {
            key: "expression",
            label: "Simple",
            options: columnNames,
            placeholder: "Column from model",
          },
          {
            key: "expression",
            label: "Advanced",
          },
        ],
        selected: item?.has("expression") ? 1 : 0,
      },
      {
        label: "Name",
        fields: [
          {
            key: "name",
            hint: "A stable identifier for use in code",
            label: "Name",
          },
        ],
        selected: 0,
      },
      {
        optional: true,
        label: "Label",
        fields: [
          {
            key: "label",
            hint: "Used on dashboards and charts",

            label: "Label",
          },
        ],
        selected: 0,
      },
      {
        label: "Format",
        fields: [
          {
            label: "Simple",
            key: "format_preset",
            options: Object.values(FormatPreset),
            placeholder: "Select a format",
          },
          {
            label: "d3-format",
            key: "format_d3",
          },
        ],
        selected: item?.get("format_d3") ? 1 : 0,
      },
      {
        optional: true,
        label: "Description",
        fields: [
          {
            key: "description",

            label: "Description",
          },
        ],
        selected: 0,
      },
      {
        label: "Summable metric",
        optional: true,
        fields: [
          {
            key: "valid_percent_of_total",
            hint: "Values can be added to yield a valid sum",
            label: "Summable metric",
            boolean: true,
          },
        ],
        selected: 0,
      },
    ],
    dimensions: [
      {
        label: "SQL Expression",
        fields: [
          {
            placeholder: "Column from model",
            options: columnNames,
            key: "column",
            label: "Simple",
          },
          {
            label: "Advanced",
            key: "expression",
          },
        ],
        selected: item?.has("column") ? 0 : 1,
      },
      {
        label: "Name",
        fields: [
          {
            key: "name",
            hint: "A stable identifier for use in code",
            label: "Name",
          },
        ],
        selected: 0,
        optional: true,
      },
      {
        label: "Label",
        optional: true,
        fields: [
          {
            key: "label",
            hint: "Used on dashboards and charts",

            label: "Label",
          },
        ],
        selected: 0,
      },

      {
        label: "Description",
        optional: true,
        fields: [
          {
            key: "description",

            label: "Description",
          },
        ],
        selected: 0,
      },
    ],
  };

  $: ({ remoteContent, localContent, saveContent } = fileArtifact);

  $: adding = index === -1;

  $: readyToGo = properties[type].every(
    ({ fields, label, optional, selected }) => {
      console.log(label, optional, editingClone[fields[selected].key]);
      return optional || editingClone[fields[selected].key];
    },
  );

  $: unsavedChanges = Object.keys(editingClone).some(
    (key) => editingClone[key] !== item?.get(key),
  );

  async function saveChanges() {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const items = (parsedDocument.get(type) as YAMLSeq).items as Array<YAMLMap>;

    const newItem = new YAMLMap();

    properties[type].forEach(({ selected, fields }) => {
      const { key } = fields[selected];
      if (editingClone[key]) newItem.set(key, editingClone[key]);
    });

    if (adding) {
      items.push(newItem);
    } else {
      items[index] = newItem;
    }

    await saveContent(parsedDocument.toString());
    editingItem.set(null);
  }

  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") onCancel(unsavedChanges);
  }}
/>

<div
  class="h-full w-[320px] bg-background flex-none p-6 flex flex-col border select-none"
>
  <h1>{adding ? "Add" : "Edit"} {type.slice(0, -1)}</h1>

  <div
    class="flex flex-col gap-y-3 w-full h-fit overflow-y-auto overflow-x-visible"
  >
    {#each properties[type] as { fields, selected, label, optional } (label)}
      {@const { hint, key, options, placeholder, boolean } = fields[selected]}
      {#if boolean}
        <div class="flex gap-x-1 items-center h-full bg-white rounded-full">
          <Switch bind:checked={editingClone[key]} id="auto-save" medium />
          <Label class="font-medium text-sm" for="auto-save">{label}</Label>
          {#if hint}
            <Tooltip location="left">
              <div class="text-gray-500">
                <InfoCircle size="13px" />
              </div>
              <TooltipContent slot="tooltip-content">
                {hint}
              </TooltipContent>
            </Tooltip>
          {/if}
        </div>
      {:else}
        <Input
          bind:value={editingClone[key]}
          {label}
          {hint}
          {optional}
          {placeholder}
          bind:selected
          fields={fields.map(({ label }) => label)}
          {options}
        />
      {/if}
    {/each}

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
            onCancel(unsavedChanges);
          }}
        >
          Cancel
        </Button>
        <Button
          type="primary"
          on:click={saveChanges}
          disabled={!readyToGo || !unsavedChanges}
        >
          {adding ? "Add " + type.slice(0, -1) : "Save changes"}
        </Button>
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
