<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ErrorsOverviewSection from "@rilldata/web-common/features/resources/overview/ErrorsOverviewSection.svelte";
  import { groupErrorsByKind } from "@rilldata/web-common/features/resources/overview-utils";
  import {
    createRuntimeServiceGetResource,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  // Resource errors grouped by kind
  $: resourcesQuery = useResources(instanceId);
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: erroredResources = allResources.filter((r) => !!r.meta?.reconcileError);
  $: errorsByKind = groupErrorsByKind(erroredResources);
  $: totalErrors = parseErrors.length + erroredResources.length;
</script>

<ErrorsOverviewSection
  parseErrorCount={parseErrors.length}
  {errorsByKind}
  {totalErrors}
  isLoading={$projectParserQuery.isLoading || $resourcesQuery.isLoading}
  isError={$projectParserQuery.isError || $resourcesQuery.isError}
  onSectionClick={() => goto(`${basePage}/resources?status=error`)}
  onParseErrorChipClick={() => goto(`${basePage}/resources?status=error`)}
  onKindChipClick={(kind) =>
    goto(`${basePage}/resources?status=error&kind=${kind}`)}
/>
