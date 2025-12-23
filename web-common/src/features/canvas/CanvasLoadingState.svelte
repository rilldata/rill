<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import DashboardBuilding from "../dashboards/DashboardBuilding.svelte";
  import DelayedSpinner from "../entity-management/DelayedSpinner.svelte";

  export let ready: boolean;
  export let errorMessage: string | undefined;
  export let isReconciling: boolean | undefined;
  export let isLoading: boolean | undefined;
</script>

<div class="size-full justify-center items-center flex flex-col">
  {#if errorMessage}
    <ErrorPage
      statusCode={404}
      header="Canvas not found"
      body={errorMessage || "An unknown error occurred."}
    />
  {:else if isReconciling}
    <DashboardBuilding />
  {:else if isLoading}
    <DelayedSpinner isLoading={true} size="48px" />
  {:else if ready}
    <slot />
  {:else}
    <header
      role="presentation"
      class="bg-background border-b py-4 px-2 w-full h-[100px] select-none z-50 flex items-center justify-center"
    ></header>
    <div class="size-full flex justify-center items-center">
      <DelayedSpinner isLoading={true} size="48px" />
    </div>
  {/if}
</div>
