<script lang="ts" context="module">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let builders: Builder[] = [];
  export let embed = false;
  // export let selected = false;
  export let interacting = false;
  // export let width: number;
  // export let height: number;
  // export let top: number;
  // export let left: number;
  export let localZIndex = 0;
  export let componentName: string;
  export let instanceId: string;

  $: resourceQuery = useResource(
    instanceId,
    componentName,
    ResourceKind.Component,
  );
  $: ({ data: componentResource } = $resourceQuery);

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});

  $: title = rendererProperties?.title;
  $: description = rendererProperties?.description;

  // FIXME: already setting staticGrid in CanvasDashboardEmbed
  // onMount(async () => {
  //   if (!embed) {
  //     gridStackManager.setStatic(true);
  //   }
  // });
</script>

<div
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  role="presentation"
  data-index={i}
  class="canvas-component hover:cursor-pointer active:cursor-grab pointer-events-auto"
  class:!cursor-default={embed}
  style:z-index={renderer === "select" ? 100 : localZIndex}
  on:contextmenu
  on:pointerenter
  on:pointerleave
>
  <div class="size-full relative">
    <div
      class="size-full overflow-hidden flex flex-col flex-none"
      class:shadow-lg={interacting}
    >
      <div class="size-full overflow-hidden flex flex-col gap-y-1 flex-none">
        {#if title || description}
          <div class="w-full h-fit flex flex-col border-b bg-white p-2">
            {#if title}
              <h1 class="text-slate-700">{title}</h1>
            {/if}
            {#if description}
              <h2 class="text-slate-600 leading-none">{description}</h2>
            {/if}
          </div>
        {/if}
        {#if renderer && rendererProperties}
          <ComponentRenderer {renderer} {componentName} />
        {/if}
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  h1 {
    font-size: 16px;
    font-weight: 500;
  }

  h2 {
    font-size: 12px;
    font-weight: 400;
  }
</style>
