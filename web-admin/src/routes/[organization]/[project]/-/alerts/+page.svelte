<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ContentContainer>
  <div class="flex flex-col items-center">
    Alerts for {organization}/{project}
  </div>
</ContentContainer>
