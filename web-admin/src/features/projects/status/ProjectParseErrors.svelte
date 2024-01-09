<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceGetResource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: parserErrors = createRuntimeServiceGetResource(
    $runtime.instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    },
    {
      query: {
        select: (data) => data.resource.projectParser.state.parseErrors,
      },
    },
  );
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">Parse errors</h2>

  {#if $parserErrors.isLoading}
    <Spinner status={EntityStatus.Running} size={"16px"} />
  {:else if $parserErrors.error}
    <div class="text-red-500">
      Error loading parse errors: {$parserErrors.error?.message}
    </div>
  {:else if $parserErrors.isSuccess}
    {#if !$parserErrors.data || $parserErrors.data.length === 0}
      <div class="text-gray-600">None!</div>
    {:else}
      <ul class="border rounded-sm">
        {#each $parserErrors.data as error}
          <li
            class="flex gap-x-5 justify-between py-1 px-9 border-b border-gray-200 bg-red-50 font-mono last:border-b-0"
          >
            <span class="text-red-600 break-all">
              {error.message}
            </span>
            {#if error.filePath}
              <span class="text-stone-500 font-semibold shrink-0">
                {error.filePath}
              </span>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>
