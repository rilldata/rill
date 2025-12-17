<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import { Copy, CheckIcon } from "lucide-svelte";
  import type { EnvVariable } from "./types";

  export let open = false;
  export let currentVariables: EnvVariable[] = [];

  let copied = false;
  let copiedTimeout: ReturnType<typeof setTimeout>;

  const pullCommand = "rill env pull";

  function handleCopyCommand() {
    navigator.clipboard.writeText(pullCommand);
    copied = true;
    clearTimeout(copiedTimeout);
    copiedTimeout = setTimeout(() => {
      copied = false;
    }, 2000);
    }
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Pull Environment Variables</DialogTitle>
      <DialogDescription>
        Pull environment variables from your Rill Cloud project to your local
        .env file.
      </DialogDescription>
    </DialogHeader>
    <div class="py-4 space-y-4">
      <p class="text-sm text-gray-700">
        Use the Rill CLI to pull environment variables from your cloud project.
        This will merge cloud variables with your local .env file.
      </p>

      <div class="space-y-2">
        <p class="text-xs font-medium text-gray-600 uppercase">Command</p>
        <div class="relative">
          <div class="bg-gray-50 border border-gray-200 rounded-md p-3 pr-12 font-mono text-sm">
            {pullCommand}
          </div>
          <button
            class="absolute right-2 top-1/2 -translate-y-1/2 p-2 hover:bg-gray-100 rounded transition-colors"
            on:click={handleCopyCommand}
            aria-label="Copy command"
          >
            {#if copied}
              <CheckIcon size="16px" class="text-green-600" />
            {:else}
              <Copy size="16px" class="text-gray-600" />
            {/if}
          </button>
        </div>
      </div>

      <div class="space-y-2">
        <p class="text-xs font-medium text-gray-600 uppercase">Options</p>
        <ul class="text-sm text-gray-600 space-y-1">
          <li><code class="bg-gray-100 px-1 rounded">--environment dev</code> - Pull dev environment variables (default)</li>
          <li><code class="bg-gray-100 px-1 rounded">--environment prod</code> - Pull prod environment variables</li>
          <li><code class="bg-gray-100 px-1 rounded">--project &lt;name&gt;</code> - Specify project name</li>
        </ul>
      </div>

      <p class="text-xs text-gray-500">
        You currently have <strong>{currentVariables.length}</strong> variable{currentVariables.length ===
        1
          ? ""
          : "s"} in your local .env file.
      </p>
    </div>
    <DialogFooter>
      <Button type="plain" onClick={() => (open = false)}>Close</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
