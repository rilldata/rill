<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";

  export let command: string;

  let copied = false;

  function copyCommand() {
    copyToClipboard(command);
    copied = true;
    setTimeout(() => (copied = false), 2500);
  }
</script>

<div class="flex flex-row text-gray-800 my-1">
  <div
    class="p-0.5 border border-gray-200 rounded-bl-sm rounded-tl-sm bg-gray-50 command-text"
  >
    {command}
  </div>
  <div
    class="p-1 border border-gray-200 rounded-br-sm rounded-tr-sm bg-gray-50
    {copied ? '' : 'cursor-pointer'}"
    on:click={copyCommand}
    on:keydown={copyCommand}
  >
    {#if copied}
      <Check size="16px" />
    {:else}
      <CopyIcon color="gray-400" size="14px" />
    {/if}
  </div>
</div>

<style>
  .command-text {
    font-family: "Source Code Variable", monospace;
    padding-left: 7px;
    padding-right: 7px;
    border-right: 0;
  }
</style>
