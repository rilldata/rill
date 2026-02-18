<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import PreviewPlaceholder from "@rilldata/web-common/components/preview/PreviewPlaceholder.svelte";
  import { SettingsIcon } from "lucide-svelte";
</script>

<ContentContainer title="Settings">
  <PreviewPlaceholder
    title="Settings are available on Rill Cloud"
    description="Deploy your project to manage environment variables, public URLs, and more"
  >
    <SettingsIcon slot="icon" size="32px" class="text-fg-muted" />
  </PreviewPlaceholder>
</ContentContainer>
