<script lang="ts">
  import type { Snippet } from "svelte";
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let { children }: { children: Snippet } = $props();

  const runtimeClient = useRuntimeClient();

  // Prefetch connectors and load into cache, but do not show a spinner,
  // it can be a bit jarring to see a lot of spinners
  const connectorsQuery = createRuntimeServiceAnalyzeConnectors(
    runtimeClient,
    {},
  );
</script>

<!-- Trigger load with hidden div so query is fired -->
<div class="hidden">${$connectorsQuery.data}</div>

<div class="flex size-full overflow-hidden">
  <div class="scroll">
    <div class="wrapper column p-10 2xl:py-16">
      <div class="mx-auto my-auto">
        {@render children()}
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .scroll {
    @apply size-full overflow-x-hidden overflow-y-auto;
  }

  .wrapper {
    @apply w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .wrapper {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
