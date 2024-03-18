<script lang="ts">
  import { chartPromptsStore } from "@rilldata/web-common/features/charts/prompt/chartPrompt";

  export let entityName: string;

  $: chartPromptHistory = chartPromptsStore.getHistoryForEntity(entityName);
  $: console.log(entityName, $chartPromptHistory);
</script>

{#if $chartPromptHistory.length}
  <div class="flex flex-col py-2">
    <div class="text-sm">Prompt history:</div>
    {#each $chartPromptHistory as history}
      <div class="text-xs">{history.prompt}</div>
    {/each}
  </div>
{/if}
