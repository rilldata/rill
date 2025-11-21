<script lang="ts">
  import { page } from "$app/stores";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import { isEmbedPage } from "@rilldata/web-common/layout/navigation/navigation-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { derived } from "svelte/store";
  import { useTransformCanvasToExploreState } from "./canvas-explore-transformer";
  import ExploreLink from "@rilldata/web-common/features/explores/explore-link/ExploreLink.svelte";

  export let component: BaseCanvasComponent<ComponentWithMetricsView>;
  export let mode: "inline" | "dropdown-item" | "icon-button" = "inline";

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: spec = component.specStore;
  $: metricsViewName = $spec?.metrics_view;

  $: isEmbedded = isEmbedPage($page);

  // Check if component can be linked to explore
  $: exploreAvailability = useExploreAvailability(instanceId, metricsViewName);

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
    {organization}
    {project}
    {exploreState}
    {isEmbedded}
    {mode}
  />
{/if}
