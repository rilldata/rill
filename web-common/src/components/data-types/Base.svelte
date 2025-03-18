<script lang="ts">
  import { onMount, onDestroy } from "svelte";

  export let classes = "";
  export let isNull = false;
  export let dark = false;
  export let truncate = false;
  export let color = "text-gray-900";
  export let contentRect: DOMRect | undefined = undefined;

  $: color = dark ? "" : color;

  let element: HTMLSpanElement;
  onMount(() => {
    if (element) {
      contentRect = element.getBoundingClientRect();
    }
  });

  onDestroy(() => {
    contentRect = undefined;
  });
</script>

<span
  bind:this={element}
  class:truncate
  class:inline-block={!truncate}
  class="whitespace-nowrap {classes} {color} break-normal pointer-events-none"
>
  {#if isNull}
    <span class="text-gray-400">-</span>
  {:else}
    <slot />
  {/if}
</span>
