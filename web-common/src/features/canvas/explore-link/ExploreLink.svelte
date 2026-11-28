<script lang="ts">
  import { page } from "$app/stores";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { useTransformCanvasToExploreState } from "./canvas-explore-transformer";
  import ExploreLink from "@rilldata/web-common/features/explores/explore-link/ExploreLink.svelte";

  const client = useRuntimeClient();

  export let component: BaseCanvasComponent<ComponentWithMetricsView>;
  export let mode: "inline" | "dropdown-item" | "icon-button" = "inline";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: spec = component.viewSpecStore;
  $: metricsViewName = $spec?.metrics_view ?? "";

  // Check if component can be linked to explore
  $: exploreAvailability = useExploreAvailability(client, metricsViewName);

  // Built when the link is clicked:
  // components read their own view state (a pivot's rows, columns and sort) with `get`,
  // so a value computed here reactively would keep that state as it was at the last time or filter change.
  // For a pivot in a canvas tab, that is before the pivot has seeded its rows and columns.
  function getExploreState() {
    return useTransformCanvasToExploreState(
      component,
      get(component.timeAndFilterStore),
    );
  }
</script>

{#if $exploreAvailability.isAvailable}
  <ExploreLink
    exploreName={$exploreAvailability.exploreName ?? metricsViewName}
    displayName={$exploreAvailability.displayName}
    {organization}
    {project}
    exploreState={getExploreState}
    {mode}
  />
{/if}
