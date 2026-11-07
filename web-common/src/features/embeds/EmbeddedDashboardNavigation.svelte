<script lang="ts">
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import TwoTieredBreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/TwoTieredBreadcrumbItem.svelte";
  import type { PathOptions } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { DashboardShortcutAction } from "@rilldata/web-common/features/dashboards/shortcuts/dashboard-shortcuts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let navigationEnabled = true;
  export let navigationBarEnabled = true;
  export let onProjectPage = false;
  export let currentResourceName: string | undefined = undefined;
  export let breadcrumbOptions: PathOptions;
  export let twoTiered = false;

  $: pathOptions = {
    ...breadcrumbOptions,
    shortcut: DashboardShortcutAction.OpenDashboardList,
  };
</script>

{#if navigationEnabled}
  <nav class="flex-1" class:sr-only={!navigationBarEnabled}>
    <ol class="flex items-center pl-4">
      {#if !onProjectPage}
        <div class="flex gap-x-2">
          <a
            class="text-fg-muted hover:text-fg-secondary"
            href="/-/embed"
            data-dashboard-shortcut={DashboardShortcutAction.GoHome}
          >
            {m.embed_home()}
          </a>
          <span class="text-fg-muted">/</span>
        </div>
      {/if}
      {#if currentResourceName}
        {#if twoTiered}
          <TwoTieredBreadcrumbItem
            {pathOptions}
            current={currentResourceName}
            isCurrentPage
          />
        {:else}
          <BreadcrumbItem
            {pathOptions}
            current={currentResourceName}
            isCurrentPage
            isEmbedded
          />
        {/if}
      {/if}
    </ol>
  </nav>
{:else}
  <div class="flex-1"></div>
{/if}
