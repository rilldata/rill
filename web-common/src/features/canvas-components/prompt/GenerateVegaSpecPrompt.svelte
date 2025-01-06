<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import ChartPromptHistoryDisplay from "@rilldata/web-common/features/canvas-components/prompt/ChartPromptHistoryDisplay.svelte";
  import { createChartGenerator } from "@rilldata/web-common/features/canvas-components/prompt/generateChart";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let open: boolean;
  export let chart: string;
  export let filePath: string;

  let prompt: string;

  $: ({ instanceId } = $runtime);

  $: generateVegaConfig = createChartGenerator(instanceId, chart, filePath);

  async function createVegaConfig() {
    open = false;
    await generateVegaConfig(prompt);
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Title>Generate vega config using AI</Dialog.Title>

    <Input bind:value={prompt} label="Prompt" />
    <ChartPromptHistoryDisplay
      entityName={chart}
      on:reuse-prompt={({ detail }) => {
        prompt = detail;
      }}
    />
    <div class="ml-auto">
      <Button on:click={createVegaConfig} large type="primary">Generate</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
