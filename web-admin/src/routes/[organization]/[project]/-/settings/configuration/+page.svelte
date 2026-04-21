<script lang="ts">
  import CodeBlock from "@rilldata/web-common/components/code-block/CodeBlock.svelte";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  $: fileQuery = createRuntimeServiceGetFile(
    runtimeClient,
    { path: "/rill.yaml" },
    {
      query: {
        retry: false,
      },
    },
  );

  $: fileContent = ($fileQuery.data?.blob as string) ?? "";
  $: hasFile = $fileQuery.isSuccess && fileContent.trim().length > 0;
</script>

<div class="flex flex-col gap-6 w-full overflow-hidden">
  <div class="flex flex-col">
    <RadixLarge>Project configuration</RadixLarge>
    <p class="text-sm text-fg-tertiary font-medium">
      Project-level settings defined in <code>rill.yaml</code>, including
      defaults, variables, and feature flags.
      <a
        href="https://docs.rilldata.com/reference/project-files/rill-yaml"
        target="_blank"
        class="text-primary-600 hover:text-primary-700 active:text-primary-800"
      >
        Learn more ->
      </a>
    </p>
  </div>

  {#if $fileQuery.isLoading}
    <DelayedSpinner isLoading={$fileQuery.isLoading} size="1rem" />
  {:else if $fileQuery.isError || !hasFile}
    <div
      class="flex items-center justify-center border rounded-sm bg-surface-subtle text-fg-tertiary text-sm py-10"
    >
      This project has no <code class="mx-1">rill.yaml</code> file.
    </div>
  {:else}
    <CodeBlock code={fileContent} language="yaml" />
  {/if}
</div>
