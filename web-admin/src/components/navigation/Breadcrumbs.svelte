<script lang="ts">
  import { page } from "$app/stores";
  import Slash from "@rilldata/web-common/components/icons/Slash.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: dashboard = $page.params.dashboard;

  const textClasses = "px-1 text-black font-semibold";
</script>

<div class="flex flex-row items-center">
  {#if organization}
    <a href="/-/{organization}" class={textClasses}>{organization}</a>
  {/if}
  {#if project}
    <Slash size={"2em"} className={"text-gray-200"} />
    <a href="/-/{organization}/{project}" class={textClasses}>{project}</a>
  {/if}
  {#if dashboard}
    <Slash size={"2em"} className={"text-gray-200"} />
    <a href="/-/{organization}/{project}/{dashboard}" class={textClasses}>
      {dashboard}
    </a>
  {/if}
</div>
