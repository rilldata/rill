<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import APIMetadata from "@rilldata/web-admin/features/apis/metadata/APIMetadata.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: api = $page.params.api;
</script>

<ContentContainer maxWidth={1100}>
  <div class=" flex flex-col items-start gap-y-9">
    <APIMetadata {api} {organization} {project} />
  </div>
</ContentContainer>
