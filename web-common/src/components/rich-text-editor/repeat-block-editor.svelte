<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricsViewName: string;
  export let canvasName: string;
  export let measure: string = "";
  export let dimension: string = "";
  export let orderBy: string = "";
  export let limit: number | undefined = undefined;
  export let where: string = "";
  export let onSave: (config: {
    measure: string;
    dimension: string;
    orderBy?: string;
    limit?: number;
    where?: string;
  }) => void = () => {};
  export let onCancel: () => void = () => {};

  $: ({ instanceId } = $runtime);
  $: ctx = getCanvasStore(canvasName, instanceId);

  let localMeasure = measure;
  let localDimension = dimension;
  let localOrderBy = orderBy;
  let localLimit = limit?.toString() ?? "";
  let localWhere = where;

  function handleSave() {
    onSave({
      measure: localMeasure,
      dimension: localDimension,
      orderBy: localOrderBy || undefined,
      limit: localLimit ? parseInt(localLimit, 10) : undefined,
      where: localWhere || undefined,
    });
  }
</script>

<div class="repeat-block-editor p-4 border border-gray-300 rounded bg-white shadow-lg">
  <div class="mb-4">
    <h3 class="text-lg font-semibold mb-2">Configure Repeat Block</h3>
    <p class="text-sm text-gray-600">
      Loop over data from {metricsViewName}
    </p>
  </div>

  <div class="space-y-4">
    <div>
      <SingleFieldInput
        {canvasName}
        label="Measure"
        metricName={metricsViewName}
        id="repeat-measure"
        type="measure"
        selectedItem={localMeasure}
        onSelect={(field) => {
          localMeasure = field;
        }}
      />
    </div>

    <div>
      <SingleFieldInput
        {canvasName}
        label="Dimension"
        metricName={metricsViewName}
        id="repeat-dimension"
        type="dimension"
        selectedItem={localDimension}
        onSelect={(field) => {
          localDimension = field;
        }}
      />
    </div>

    <div>
      <InputLabel small label="ORDER BY (optional)" id="repeat-orderby" />
      <Input
        inputType="text"
        size="sm"
        placeholder="e.g., revenue DESC"
        bind:value={localOrderBy}
      />
      <p class="text-xs text-gray-500 mt-1">
        Column name, optionally followed by DESC or ASC
      </p>
    </div>

    <div>
      <InputLabel small label="LIMIT (optional)" id="repeat-limit" />
      <Input
        inputType="number"
        size="sm"
        placeholder="e.g., 100"
        bind:value={localLimit}
      />
    </div>

    <div>
      <InputLabel small label="WHERE (optional)" id="repeat-where" />
      <textarea
        class="w-full p-2 border border-gray-300 rounded-sm text-sm"
        rows="3"
        placeholder='e.g., country != "US"'
        bind:value={localWhere}
      />
      <p class="text-xs text-gray-500 mt-1">
        SQL WHERE clause conditions
      </p>
    </div>
  </div>

  <div class="flex gap-2 mt-6 justify-end">
    <Button type="subtle" onClick={onCancel}>Cancel</Button>
    <Button
      onClick={handleSave}
      disabled={!localMeasure || !localDimension}
    >
      Save
    </Button>
  </div>
</div>

<style lang="postcss">
  .repeat-block-editor {
    min-width: 400px;
    max-width: 600px;
  }
</style>

