<script lang="ts">
  import { page } from "$app/stores";
  import Slash from "@rilldata/web-common/components/icons/Slash.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: name = $page.params.name;

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
  {#if name}
    <Slash size={"2em"} className={"text-gray-200"} />
    <a href="/-/{organization}/{project}/dashboard/{name}" class={textClasses}>
      {name}
    </a>
  {/if}
</div>
