<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  // Temporary page for testing

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: alerts = useAlerts($runtime.instanceId);
</script>

<ContentContainer>
  <div class="flex flex-col items-center">
    Alerts for {organization}/{project}

    {#if $alerts.data?.resources}
      {#each $alerts.data.resources as alert}
        <a href={`alerts/${alert.meta.name.name}`}>{alert.meta.name.name}</a>
      {/each}
    {/if}
  </div>
</ContentContainer>
