<script lang="ts">
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let command: string;
  export let className = "";

  let copied = false;

  function copyCommand() {
    copyToClipboard(command);
    copied = true;
    setTimeout(() => (copied = false), 2500);
  }
</script>

<div class="flex flex-row text-gray-800 my-1 {className}">
  <div class="p-0.5 border rounded-bl-sm rounded-tl-sm bg-gray-50 command-text">
    {command}
  </div>
  <button
    type="button"
    class="p-1 border rounded-br-sm rounded-tr-sm bg-gray-50 cursor-pointer"
    on:click={copyCommand}
  >
    {#if copied}
      <Check size="14px" />
    {:else}
      <CopyIcon color="gray-400" size="14px" />
    {/if}
  </button>
</div>

<style>
  .command-text {
    font-family: "Source Code Variable", monospace;
    padding-left: 7px;
    padding-right: 7px;
    border-right: 0;
  }
</style>
