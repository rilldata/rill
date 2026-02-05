<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ResourceError from "@rilldata/web-admin/features/projects/ResourceError.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1ListResourcesResponse } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { CreateQueryResult, QueryKey } from "@tanstack/svelte-query";

  export let kind: "report" | "dashboard" | "alert";
  export let query: CreateQueryResult<V1ListResourcesResponse, HTTPError> & {
    queryKey: QueryKey;
  };

  $: ({ isLoading, isError, isSuccess, error } = $query);
</script>

<ContentContainer title="Project {kind}s">
  <div class="flex flex-col items-center gap-y-4">
    {#if isLoading}
      <div class="m-auto mt-20">
        <DelayedSpinner isLoading size="24px" />
      </div>
    {:else if isError}
      <ResourceError {kind} {error} />
    {:else if isSuccess}
      <slot name="table" />
    {/if}
  </div>
</ContentContainer>
