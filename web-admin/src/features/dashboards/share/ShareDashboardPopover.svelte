<script lang="ts">
  import CreatePublicURLForm from "@rilldata/web-admin/features/public-urls/CreatePublicURLForm.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import ExportDashboardForm from "@rilldata/web-common/features/exports/pdf/ExportDashboardForm.svelte";
  import { exportCanvasPdf } from "@rilldata/web-common/features/exports/pdf/export-canvas-pdf";
  import type { PdfExportRunOptions } from "@rilldata/web-common/features/exports/pdf/types";

  export let createMagicAuthTokens: boolean;
  // Provide canvas identifiers to enable the "PDF" tab (canvas dashboards only).
  export let canvasName: string | undefined = undefined;
  export let instanceId: string | undefined = undefined;

  const { hidePublicUrl } = featureFlags;
  let isOpen = false;
  let copied = false;
  let runPdfExport: ((o: PdfExportRunOptions) => Promise<void>) | null = null;

  // Bind the (now-narrowed) identifiers in a helper so the returned closure keeps
  // them as `string` rather than `string | undefined`.
  $: runPdfExport =
    canvasName && instanceId ? makeRunPdfExport(canvasName, instanceId) : null;

  function makeRunPdfExport(name: string, id: string) {
    return (o: PdfExportRunOptions) =>
      exportCanvasPdf({ canvasName: name, instanceId: id, ...o });
  }

  function onCopy() {
    navigator.clipboard.writeText(window.location.href).catch(console.error);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger>
    {#snippet child({ props })}
      <Tooltip distance={8} suppress={isOpen}>
        <Button {...props} type="secondary" selected={isOpen}>Share</Button>
        <TooltipContent slot="tooltip-content">Share dashboard</TooltipContent>
      </Tooltip>
    {/snippet}
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">Copy URL</TabsTrigger>
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <TabsTrigger value="tab2">Create public URL</TabsTrigger>
        {/if}
        {#if runPdfExport}
          <TabsTrigger value="pdf">PDF</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-fg-primary font-normal">
            Share your current view with another project member.
          </h3>
          <Button
            type="secondary"
            onClick={() => {
              onCopy();
            }}
          >
            {#if copied}
              <Check size="16px" />
              Copied URL
            {:else}
              <Link size="16px" className="text-primary-500" />
              Copy URL for this view
            {/if}
          </Button>
        </div>
      </TabsContent>
      <TabsContent value="tab2" class="mt-0 p-4">
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <CreatePublicURLForm />
        {/if}
      </TabsContent>
      {#if runPdfExport}
        <TabsContent value="pdf" class="mt-0 p-4">
          <ExportDashboardForm
            runExport={runPdfExport}
            onComplete={() => (isOpen = false)}
          />
        </TabsContent>
      {/if}
    </Tabs>
  </PopoverContent>
</Popover>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
