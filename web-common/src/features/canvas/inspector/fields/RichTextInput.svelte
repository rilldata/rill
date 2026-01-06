<script lang="ts">
  import RichTextEditor from "@rilldata/web-common/components/rich-text-editor/RichTextEditor.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { derived } from "svelte/store";

  import { onMount } from "svelte";

  export let component: BaseCanvasComponent;
  export let label: string;
  export let id: string;
  export let hint: string | undefined = undefined;
  export let value: string = "";
  export let onUpdate: (value: string) => void = () => {};

  $: ({ instanceId } = $runtime);
  $: ctx = getCanvasStore(component.parent.name, instanceId);

  // Get metrics view from parent canvas
  $: parentSpecStore = component.parent.specStore;
  $: parentSpec = $parentSpecStore;
  $: metricsViews = parentSpec?.data?.metricsViews ?? {};

  // Get the first metrics view name and its measures
  $: metricsViewName = Object.keys(metricsViews)[0];
  $: metricsView = metricsViewName ? metricsViews[metricsViewName] : undefined;
  $: availableMeasures =
    metricsView?.state?.validSpec?.measures?.map((m) => m.name) ?? [];

  let content = value;

  $: if (value !== content) {
    content = value;
  }

  function handleUpdate(newContent: string) {
    content = newContent;
    onUpdate(newContent);
  }
</script>

  <div class="flex flex-col gap-y-2">
  <InputLabel {hint} small {label} {id} />
  {#if metricsViewName}
    <RichTextEditor
      bind:content
      {metricsViewName}
      {availableMeasures}
      placeholder={label}
      onUpdate={handleUpdate}
    />
  {:else}
    <RichTextEditor
      bind:content
      placeholder={label}
      onUpdate={handleUpdate}
    />
  {/if}
</div>

