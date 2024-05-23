<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";

  export let markdown: string = "";
  export let editing = false;
  export let css = {};
  export let onChange:
    | undefined
    | ((
        e: Event & {
          currentTarget: EventTarget & HTMLInputElement;
        },
      ) => void) = undefined;

  $: styleString = Object.entries(css)
    .map(([k, v]) => `${k}:${v}`)
    .join(";");
</script>

<div
  class="markdown size-full items-center flex justify-center"
  style={styleString}
>
  {#if editing}
    <input
      class="w-full bg-transparent"
      type="text"
      value={markdown}
      on:input={onChange}
    />
  {:else}
    {#await marked(markdown) then content}
      {@html DOMPurify.sanitize(content)}
    {/await}
  {/if}
</div>
