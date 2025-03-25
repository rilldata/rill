<script lang="ts">
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { QueryObserverResult } from "@tanstack/svelte-query";
  import { onDestroy } from "svelte";

  export let queries: {
    label: string;
    query: QueryObserverResult<unknown, HTTPError>;
  }[];
  export let longLoadThreshold: number;

  $: loading = queries.some(({ query }) => query.isLoading);
  let loadingForLong = false;

  $: errors = queries
    .filter(({ query }) => !!query.error)
    .map(({ label, query }) => ({ label, error: query.error }));

  let timeoutId: ReturnType<typeof setTimeout>;
  $: {
    clearTimeout(timeoutId);
    if (loading) {
      timeoutId = setTimeout(() => (loadingForLong = true), longLoadThreshold);
    } else {
      loadingForLong = false;
    }
  }
  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if loading}
  <slot name="loading" {loadingForLong} />
{:else if errors.length > 0}
  <slot name="errors" {errors} />
{:else}
  <slot />
{/if}
