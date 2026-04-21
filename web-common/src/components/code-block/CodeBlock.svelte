<script lang="ts">
  import "prismjs/themes/prism.css";
  import { onMount } from "svelte";
  import Button from "../button/Button.svelte";

  export let code: string;
  export let language: string = "json";
  export let showCopyButton: boolean = true;

  let codeElement: HTMLElement;
  let copied = false;
  let ready = false;

  // Prism's language plugins reference bare `Prism` as a global. Rolldown may
  // bundle them into a separate chunk that evaluates before the main module
  // has set window.Prism. Dynamic imports in onMount guarantee correct ordering.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let Prism: any;

  function copyCode() {
    navigator.clipboard.writeText(code);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }

  onMount(async () => {
    const mod = await import("prismjs");
    Prism = mod.default;
    (window as any).Prism = Prism; // eslint-disable-line @typescript-eslint/no-explicit-any
    await import("prismjs/components/prism-json");
    ready = true;
  });

  $: if (ready && codeElement && code !== undefined && language !== undefined) {
    Prism.highlightElement(codeElement);
  }
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
    @apply bg-surface-subtle;
    border-radius: 4px;
    padding: 1em;
    overflow: auto;
    margin: 0;
  }

  code {
    text-shadow: none;
  }

  :global(.operator) {
    background: none !important;
  }
</style>
