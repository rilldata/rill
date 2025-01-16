<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import NoResourceCTA from "@rilldata/web-admin/features/projects/NoResourceCTA.svelte";
  import ResourceError from "@rilldata/web-admin/features/projects/ResourceError.svelte";
  import type { CreateQueryResult, QueryKey } from "@tanstack/svelte-query";
  import type { V1ListResourcesResponse } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

  export let kind: "report" | "dashboard" | "alert";
  export let query: CreateQueryResult<V1ListResourcesResponse, HTTPError> & {
    queryKey: QueryKey;
  };

  $: ({ data, isLoading, isError, isSuccess } = $query);

  $: resources = data?.resources ?? [];
</script>

<ContentContainer title="Project {kind}s" showTitle={!!resources.length}>
  <div class="flex flex-col items-center gap-y-4">
    {#if isLoading}
      <div class="m-auto mt-20">
        <DelayedSpinner isLoading size="24px" />
      </div>
    {:else if isError}
      <ResourceError {kind} />
    {:else if isSuccess}
      {#if !resources?.length}
        <NoResourceCTA {kind}>
          <slot name="action" />
        </NoResourceCTA>
      {:else}
        <slot name="table" />
      {/if}
    {/if}
  </div>
</ContentContainer>
