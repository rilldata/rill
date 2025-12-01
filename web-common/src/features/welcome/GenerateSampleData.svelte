<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { generateModel } from "@rilldata/web-common/features/chat/core/actions.ts";

  export let isInit: boolean;
  export let open = false;

  $: ({ instanceId } = $runtime);
  let prompt = "";

  async function initProjectWithSampleData() {
    void generateModel(
      isInit,
      instanceId,
      `Generate a model for the following user prompt: ${prompt}`,
    );
    open = false;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild let:builder>
    {#if isInit}
      <Button type="ghost" builders={[builder]} large>
        or generate sample data using AI
      </Button>
    {:else}
      <div class="hidden"></div>
    {/if}
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Generate sample data</Dialog.Title>
      <Dialog.Description>
        <div>What is the business context or domain of your data?</div>
      </Dialog.Description>
    </Dialog.Header>
    <textarea
      class="prompt-input"
      bind:value={prompt}
      class:empty={prompt.length === 0}
      placeholder="e.g. sales transaction of an e-commerce store"
    />
    <Button type="primary" large onClick={initProjectWithSampleData}>
      Generate
    </Button>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .prompt-input {
    @apply w-full p-2 min-h-[2.5rem];
    @apply border border-gray-300 rounded-[2px];
    @apply text-sm leading-relaxed;
  }

  .prompt-input.empty::before {
    content: attr(data-placeholder);
    @apply text-gray-400 pointer-events-none absolute;
  }
</style>
