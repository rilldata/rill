<script lang="ts">
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import PreviewPlaceholder from "@rilldata/web-common/components/preview/PreviewPlaceholder.svelte";
</script>

<svelte:head>
  <title>Alerts - Rill</title>
</svelte:head>

<ContentContainer title="Alerts">
  <PreviewPlaceholder
    title="You don't have any alerts yet"
    description="Create alerts from any dashboard or via code after deploying"
  >
    <AlertIcon slot="icon" size="32px" />
  </PreviewPlaceholder>
</ContentContainer>
