<script lang="ts">
  import { page } from "$app/stores";
  import { type EnvironmentTypes } from "@rilldata/web-admin/features/projects/environment-variables/types";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { onMount } from "svelte";
  import EnvironmentVariablesEditor from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesEditor.svelte";

  // Filters — synced to URL params `q` and `env` (multi-select array)
  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "env", type: "array" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let envFilter: EnvironmentTypes[] = parseArrayParam(
    $page.url.searchParams.get("env"),
  ) as EnvironmentTypes[];
  let mounted = false;

  // URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    searchText = parseStringParam($page.url.searchParams.get("q"));
    envFilter = parseArrayParam(
      $page.url.searchParams.get("env"),
    ) as EnvironmentTypes[];
  }

  // Local state → URL
  $: if (mounted) {
    filterSync.syncToUrl({ q: searchText, env: envFilter });
  }

  onMount(() => {
    mounted = true;
  });
</script>

<EnvironmentVariablesEditor bind:searchText bind:envFilter />
