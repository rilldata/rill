<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { chartPromptsStore } from "@rilldata/web-common/features/charts/prompt/chartPrompt";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { createEventDispatcher } from "svelte";

  export let entityName: string;

  const dispatch = createEventDispatcher();

  $: chartPromptHistory = chartPromptsStore.getHistoryForEntity(entityName);
</script>

{#if $chartPromptHistory.length}
  <div class="flex flex-col py-2">
    <div class="text-sm">Prompt history:</div>
    {#each $chartPromptHistory as history}
      <div class="flex flex-row items-center text-xs">
        <Button
          type="secondary"
          compact
          noStroke
          on:click={() => dispatch("reuse-prompt", history.prompt)}
        >
          <RefreshIcon size="14px" />
        </Button>
        <span>{history.prompt}</span>
      </div>
    {/each}
  </div>
{/if}
