<script lang="ts">
  export let organization: string;
  export let project: string;

  // PROBLEM: `useDashboardNames` calls a runtime, but we're currently constrained to 1 runtime per page
  // Problem detailed in bullet point #1 here: https://github.com/rilldata/rill-developer/pull/1826#issue-1600724370
  // Solution proposed here: https://github.com/rilldata/rill-developer/pull/1826#issuecomment-1455700785
  // $: dashboards = useDashboardNames($runtime?.instanceId);
</script>

<!-- {#if $dashboards.data} -->
<ol>
  <!-- {#each $dashboards.data as dashboard} -->
  {#each ["dashboard1", "dashboard2", "dashboard3"] as dashboard}
    <li class="text-xs text-gray-900 font-medium leading-4 mb-1">
      <a href="/{organization}/{project}/{dashboard}">{dashboard}</a>
    </li>
  {/each}
</ol>
<!-- {/if} -->
