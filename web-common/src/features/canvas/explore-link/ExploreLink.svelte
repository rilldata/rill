<script lang="ts">
  import { page } from "$app/stores";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { derived } from "svelte/store";
  import { useTransformCanvasToExploreState } from "./canvas-explore-transformer";
  import ExploreLink from "@rilldata/web-common/features/explores/explore-link/ExploreLink.svelte";

  const client = useRuntimeClient();

  export let component: BaseCanvasComponent<ComponentWithMetricsView>;
  export let mode: "inline" | "dropdown-item" | "icon-button" = "inline";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: spec = component.specStore;
  $: metricsViewName = $spec?.metrics_view;

  // Check if component can be linked to explore
  $: exploreAvailability = useExploreAvailability(client, metricsViewName);

  $: context = derived(
    [exploreAvailability, component.timeAndFilterStore],
    ([exploreAvailResp, timeAndFilterStore]) => ({
      organization,
      project,
      exploreName: exploreAvailResp.exploreName ?? metricsViewName,
      timeAndFilterStore,
    }),
  );

  $: exploreState = useTransformCanvasToExploreState(component, $context);
</script>

{#if $exploreAvailability.isAvailable}
  <ExploreLink
    exploreName={$context.exploreName}
    displayName={$exploreAvailability.displayName}
    {organization}
    {project}
    {exploreState}
    {mode}
  />
{/if}
