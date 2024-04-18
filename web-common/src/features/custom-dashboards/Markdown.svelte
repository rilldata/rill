<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";

  export let markdown: string;
  export let fontSize: number;
  export let editing = false;
  export let onChange:
    | undefined
    | ((
        e: Event & {
          currentTarget: EventTarget & HTMLInputElement;
        },
      ) => void) = undefined;
</script>

<div
  class="markdown size-full items-center flex justify-center"
  style:font-size="{fontSize}px"
>
  {#if editing}
    <input
      class="w-full bg-transparent"
      type="text"
      value={markdown}
      on:input={onChange}
    />
  {:else}
    {@html DOMPurify.sanitize(marked(markdown))}
  {/if}
</div>
