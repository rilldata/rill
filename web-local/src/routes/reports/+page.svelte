<script lang="ts">
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import PreviewPlaceholder from "@rilldata/web-common/components/preview/PreviewPlaceholder.svelte";
</script>

<svelte:head>
  <title>Reports - Rill</title>
</svelte:head>

<ContentContainer title="Reports">
  <PreviewPlaceholder
    title="You don't have any reports yet"
    description="Schedule reports from any dashboard after deploying"
  >
    <ReportIcon slot="icon" size="32px" />
  </PreviewPlaceholder>
</ContentContainer>
