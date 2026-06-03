<!-- A utility component that makes it simple to add a resizable sidebar to a layout.
     Maintains its own state in memory. We can expand it to be in local storage.
     TODO: replace sidebar usages across the app with this. -->
<script lang="ts" context="module">
  const widthStores = new Map<string, Writable<number>>();
  function getWidthStore(id: string, defaultWidth: number) {
    if (widthStores.has(id)) return widthStores.get(id)!;
    const store = writable(defaultWidth);
    widthStores.set(id, store);
    return store;
  }
</script>

<script lang="ts">
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { writable, type Writable } from "svelte/store";

  export let id: string;
  export let minWidth: number;
  export let defaultWidth: number;
  export let maxWidth: number;
  export let additionalClass = "";

  const store = getWidthStore(id, defaultWidth);
</script>

<div class="h-full relative {additionalClass}" style="width: {$store}px">
  <Resizer
    min={minWidth}
    max={maxWidth}
    basis={defaultWidth}
    dimension={$store}
    direction="EW"
    side="left"
    onUpdate={(d) => store.set(d)}
  />
  <slot />
</div>
