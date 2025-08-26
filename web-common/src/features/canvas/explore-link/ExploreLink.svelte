<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import { generateExploreLink } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
  import {
    ExploreLinkErrorType,
    type ExploreLinkError,
  } from "@rilldata/web-common/features/explore-mappers/types";
  import { getErrorMessage } from "@rilldata/web-common/features/explore-mappers/utils";
  import { isEmbedPage } from "@rilldata/web-common/layout/navigation/navigation-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { derived } from "svelte/store";
  import { useTransformCanvasToExploreState } from "./canvas-explore-transformer";

  export let component: BaseCanvasComponent<ComponentWithMetricsView>;
  export let mode: "inline" | "dropdown-item" | "icon-button" = "inline";

  let isNavigating = false;
  let navigationError: ExploreLinkError | null = null;

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

  async function gotoExplorePage() {
    if (
      !$exploreAvailability.isAvailable ||
      !metricsViewName ||
      !$exploreAvailability.exploreName ||
      !exploreState
    )
      return;

    navigationError = null;
    isNavigating = true;

    try {
      const exploreURL = await generateExploreLink(
        exploreState,
        $context.exploreName,
        $context.organization,
        $context.project,
      );
      await goto(exploreURL);
    } catch (error) {
      console.warn("Navigation error:", error);
      if (error.type) {
        navigationError = error as ExploreLinkError;
      } else {
        navigationError = {
          type: ExploreLinkErrorType.TRANSFORMATION_ERROR,
          message: error?.message,
          details: error,
        };
      }
    } finally {
      isNavigating = false;
    }
  }

  $: canNavigate =
    $exploreAvailability.isAvailable &&
    !isNavigating &&
    !!exploreState &&
    !isEmbedded;
</script>

{#if $exploreAvailability.isAvailable}
  {#if mode === "dropdown-item"}
    <DropdownMenu.Item on:click={gotoExplorePage}>
      {#if isNavigating}
        <Spinner status={EntityStatus.Running} size="14px" />
      {:else}
        <ExploreIcon size="14px" />
      {/if}
      Go to Explore
    </DropdownMenu.Item>
  {:else if mode === "icon-button"}
    <IconButton
      on:click={gotoExplorePage}
      size={28}
      disabled={!canNavigate}
      ariaLabel="Go to Explore Dashboard"
      disableHover={canNavigate}
    >
      {#if isNavigating}
        <Spinner status={EntityStatus.Running} size="18px" />
      {:else}
        <ExploreIcon size="18px" />
      {/if}
      <div slot="tooltip-content">Go to Explore Dashboard</div>
    </IconButton>
  {:else}
    <button
      on:click={gotoExplorePage}
      class="inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 underline cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
      disabled={!canNavigate}
      type="button"
    >
      {#if isNavigating}
        <Spinner status={EntityStatus.Running} size="1em" />
      {/if}
      Go to Explore Dashboard
    </button>
  {/if}

  {#if navigationError && mode === "inline"}
    <div class="flex flex-col gap-y-2 text-red-600 mt-2">
      <h3 class="text-sm font-semibold">Unable to open Explore Dashboard</h3>
      <p class="text-xs">{getErrorMessage(navigationError)}</p>
    </div>
  {/if}
{:else if isNavigating && mode === "inline"}
  <div class="h-36">
    <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
  </div>
{/if}
