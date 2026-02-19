<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: hasResourceErrorsQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
    {
      query: {
        select: (data) => {
          return (
            (data.resources ?? []).filter(
              (resource) => !!resource.meta?.reconcileError,
            ).length > 0
          );
        },
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );
  $: ({
    data: hasResourceErrors,
    error: hasResourceErrorsError,
    isLoading: hasResourceErrorsLoading,
  } = $hasResourceErrorsQuery);

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
  $: ({
    data: projectParserData,
    error: projectParserError,
    isLoading: projectParserLoading,
  } = $projectParserQuery);
  $: hasParseErrors =
    (projectParserData?.resource?.projectParser?.state?.parseErrors?.length ??
      0) > 0;
</script>

{#if hasResourceErrorsLoading || projectParserLoading}
  <LoadingSpinner />
{:else if hasResourceErrorsError || projectParserError}
  <CancelCircle className="text-red-600" />
{:else if hasResourceErrors || hasParseErrors}
  <CancelCircle className="text-red-600" />
{:else}
  <CheckCircle className="text-green-400" />
{/if}
