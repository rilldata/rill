<script lang="ts">
  import Prism from "prismjs";
  import "prismjs/components/prism-json";
  import "prismjs/themes/prism.css";
  import { onMount } from "svelte";
  import Button from "../button/Button.svelte";

  export let code: string;
  export let language: string = "json";
  export let showCopyButton: boolean = true;

  let codeElement: HTMLElement;
  let copied = false;

  function copyCode() {
    navigator.clipboard.writeText(code);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }

  onMount(() => {
    if (codeElement) {
      Prism.highlightElement(codeElement);
    }
  });

  $: (async () => {
    if (codeElement && code !== undefined && language !== undefined) {
      Prism.highlightElement(codeElement);
    }
  })();
</script>

<div class="relative">
  {#if showCopyButton}
    <Button
      type="secondary"
      onClick={copyCode}
      small
      class="absolute top-2 right-2 z-10"
    >
      {#if copied}Copied!{:else}Copy{/if}
    </Button>
  {/if}
  {#key code + language}
    <pre><code bind:this={codeElement} class={`language-${language}`}
        >{code}</code
      ></pre>
  {/key}
</div>

<style lang="postcss">
  pre {
    background: #f7f7f7;
    border: 1px solid #ececec;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    border-radius: 4px;
    padding: 1em;
    overflow: auto;
    margin: 0;
  }
</style>
