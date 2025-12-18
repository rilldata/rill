<script lang="ts">
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  export let title: string;
  export let yaml: string;

  let copied = false;

  function copyYaml() {
    navigator.clipboard.writeText(yaml);
    copied = true;
    setTimeout(() => (copied = false), 2_000);
  }
</script>

<div>
  <div class="text-sm leading-none font-medium mb-4">{title}</div>
  <div class="relative">
    <button
      class="absolute top-2 right-2 p-1 rounded"
      type="button"
      aria-label="Copy YAML"
      on:click={copyYaml}
    >
      {#if copied}
        <Check size="16px" />
      {:else}
        <CopyIcon size="16px" />
      {/if}
    </button>
    <pre
      class="bg-muted p-3 rounded text-xs border border-gray-200 overflow-x-auto">{yaml}</pre>
  </div>
  <slot />
  <!-- support need help, errors etc -->
</div>
