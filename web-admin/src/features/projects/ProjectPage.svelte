<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import NoResourceCTA from "@rilldata/web-admin/features/projects/NoResourceCTA.svelte";
  import ResourceError from "@rilldata/web-admin/features/projects/ResourceError.svelte";

  export let query: ReturnType<typeof useReports>;
  export let kind: "report" | "dashboard" | "alert";

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
