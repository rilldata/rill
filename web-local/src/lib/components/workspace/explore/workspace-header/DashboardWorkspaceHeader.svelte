<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import ToggleButton from "./ToggleButton.svelte";
  import ToggleButtonGroup from "./ToggleButtonGroup.svelte";
  import WorkspaceViewControls from "./WorkspaceViewControls.svelte";
  export let metricViewName;
  export let displayName = "hmm";
  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;
  $: modelLink = `/dashboard/${metricViewName}/model`;
  $: editLink = `/dashboard/${metricViewName}/edit`;
  $: dashboardLink = `/dashboard/${metricViewName}`;
  let view = "dashboard";
  // view is fast-updated in component, then ensured to be the same as the URL
  $: view =
    $page.url.pathname === modelLink
      ? "model"
      : $page.url.pathname === editLink
      ? "config"
      : $page.url.pathname === dashboardLink
      ? "dashboard"
      : "dashboard";
  $: console.log(view);
</script>

<header
  style:height="var(--header-height)"
  class="grid items-center border-b pl-2 pr-4 gap-x-4"
  style:grid-template-columns={"[title] auto [view-controls] max-content [asset-controls] max-content"}
>
  <div
    style:grid-column="title"
    style:padding-left="{$navigationVisibilityTween * 20}px"
  >
    <div class="pl-4 font-bold" style:font-size="12px">
      {displayName || metricViewName}
      <span class="text-gray-500 font-normal"> / {view}</span>
    </div>
  </div>
  <div style:grid-column="view-controls">
    {#if view !== "dashboard"}
      <WorkspaceViewControls />
    {/if}
  </div>
  <!-- top right CTAs -->
  <div style:grid-column="asset-controls" style="flex-shrink: 0;">
    <ToggleButtonGroup>
      <ToggleButton
        active={view === "model"}
        on:click={() => {
          view = "model";
          goto(modelLink);
        }}>Model</ToggleButton
      >
      <div
        style:width="1px"
        style:height="28px"
        class:bg-gray-200={view === "dashboard"}
      />
      <ToggleButton
        active={view === "config"}
        on:click={() => {
          view = "config";
          goto(editLink);
        }}>Config</ToggleButton
      >
      <div
        style:width="1px"
        style:height="28px"
        class:bg-gray-200={view === "model"}
      />
      <ToggleButton
        active={view === "dashboard"}
        on:click={() => {
          view = "dashboard";
          goto(dashboardLink);
        }}>Dashboard</ToggleButton
      >
    </ToggleButtonGroup>
  </div>
</header>
