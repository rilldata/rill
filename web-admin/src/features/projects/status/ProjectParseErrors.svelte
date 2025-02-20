<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceGetResource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: projectParserQuery = createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    },
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );
  $: ({ isLoading, isSuccess, data, error } = $projectParserQuery);

  $: parseErrors = data?.resource?.projectParser.state.parseErrors;
  $: parserReconcileError = data?.resource?.meta?.reconcileError;
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">Parse errors</h2>

  {#if isLoading}
    <Spinner status={EntityStatus.Running} size={"16px"} />
  {:else if error}
    <div class="text-red-500">
      Error loading parse errors: {error.message}
    </div>
  {:else if isSuccess}
    {#if parseErrors && parseErrors.length > 0}
      <ul class="border rounded-sm">
        {#each parseErrors as error, i (i)}
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
    {:else if parserReconcileError}
      <div class="text-red-500">
        {parserReconcileError}
      </div>
    {:else}
      <div class="text-gray-600">None!</div>
    {/if}
  {/if}
</section>
