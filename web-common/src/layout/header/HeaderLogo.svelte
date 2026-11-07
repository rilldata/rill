<script lang="ts">
  import { DashboardShortcutAction } from "@rilldata/web-common/features/dashboards/shortcuts/dashboard-shortcuts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";

  export let href: string = "/";
  export let logoUrl: string | undefined = undefined;
</script>

<a
  {href}
  data-dashboard-shortcut={DashboardShortcutAction.GoHome}
  aria-label={m.dashboard_shortcut_go_home()}
  class="grid place-content-center rounded {logoUrl ? 'pl-2 pr-2' : 'p-2'}"
>
  {#if logoUrl}
    <img src={logoUrl} alt="" class="h-7" />
  {:else}
    <Rill />
  {/if}
</a>
