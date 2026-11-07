<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import ExplainAndFixErrorButton from "@rilldata/web-common/features/chat/ExplainAndFixErrorButton.svelte";
  import {
    extractErrorMessage,
    extractErrorStatusCode,
  } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import DashboardBuilding from "../dashboards/DashboardBuilding.svelte";
  import DelayedSpinner from "../entity-management/DelayedSpinner.svelte";

  export let ready: boolean;
  export let errorMessage: string | undefined;
  export let isReconciling: boolean | undefined;
  export let isLoading: boolean | undefined;
  export let filePath: string | undefined = undefined;
  // Failure to fetch the canvas at all, as opposed to a canvas that failed to reconcile.
  export let queryError: unknown = undefined;

  $: queryErrorStatus = queryError
    ? extractErrorStatusCode(queryError)
    : undefined;
</script>

<div class="size-full justify-center items-center flex flex-col">
  {#if ready}
    <slot />
  {:else if errorMessage}
    <ErrorPage
      statusCode={404}
      header={m.canvas_not_found()}
      body={errorMessage || m.canvas_unknown_error()}
    >
      <svelte:fragment slot="cta">
        {#if filePath}
          <ExplainAndFixErrorButton {filePath} variant="cta" />
        {/if}
      </svelte:fragment>
    </ErrorPage>
  {:else if queryError}
    <ErrorPage
      statusCode={queryErrorStatus ?? 500}
      header={queryErrorStatus === 403
        ? m.error_access_denied_header()
        : m.canvas_unknown_error()}
      body={queryErrorStatus === 403
        ? m.error_access_denied_body()
        : extractErrorMessage(queryError)}
    />
  {:else if isReconciling}
    <DashboardBuilding />
  {:else if isLoading}
    <DelayedSpinner isLoading={true} size="48px" />
  {:else}
    <header
      role="presentation"
      class="bg-surface-background border-b py-4 px-2 w-full h-[100px] select-none z-50 flex items-center justify-center"
    ></header>
    <div class="size-full flex justify-center items-center">
      <DelayedSpinner isLoading={true} size="48px" />
    </div>
  {/if}
</div>
