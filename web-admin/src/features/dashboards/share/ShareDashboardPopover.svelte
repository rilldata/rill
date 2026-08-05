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
  import { getCanvasStoreUnguarded } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { LayoutBlock } from "@rilldata/web-common/features/canvas/stores/tab-group";
  import ExportDashboardForm from "@rilldata/web-common/features/exports/pdf/ExportDashboardForm.svelte";
  import { exportCanvasPdf } from "@rilldata/web-common/features/exports/pdf/export-canvas-pdf";
  import type { PdfExportRunOptions } from "@rilldata/web-common/features/exports/pdf/types";
  import ScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/ScheduledReportDialog.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { readable } from "svelte/store";

  export let createMagicAuthTokens: boolean;
  // Provide canvas identifiers to enable the "PDF" tab (canvas dashboards only).
  export let canvasName: string | undefined = undefined;
  export let instanceId: string | undefined = undefined;

  const { hidePublicUrl, reports } = featureFlags;
  let isOpen = false;
  let copied = false;
  let showScheduledReportDialog = false;
  let runPdfExport: ((o: PdfExportRunOptions) => Promise<void>) | null = null;

  // Bind the (now-narrowed) identifiers in a helper so the returned closure keeps
  // them as `string` rather than `string | undefined`.
  $: runPdfExport =
    canvasName && instanceId ? makeRunPdfExport(canvasName, instanceId) : null;

  function makeRunPdfExport(name: string, id: string) {
    return (o: PdfExportRunOptions) =>
      exportCanvasPdf({ canvasName: name, instanceId: id, ...o });
  }

  const emptyLayout = readable<LayoutBlock[]>([]);
  // Resolved when the popover opens (the canvas store may not yet be
  // initialized when this header button first mounts).
  $: layoutStore =
    (isOpen &&
      canvasName &&
      instanceId &&
      getCanvasStoreUnguarded(canvasName, instanceId)?.canvasEntity.layout) ||
    emptyLayout;
  // Gates the all-tabs/active-tab option in the PDF export form.
  $: hasTabGroups = $layoutStore.some((block) => block.kind === "tab-group");

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
        <Button {...props} type="secondary" selected={isOpen}
          >{m.avatar_share()}</Button
        >
        <TooltipContent slot="tooltip-content"
          >{m.avatar_share_dashboard()}</TooltipContent
        >
      </Tooltip>
    {/snippet}
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">{m.avatar_copy_url()}</TabsTrigger>
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <TabsTrigger value="tab2">{m.avatar_create_public_url()}</TabsTrigger>
        {/if}
        {#if runPdfExport}
          <TabsTrigger value="pdf">PDF</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-fg-primary font-normal">
            {m.avatar_share_description()}
          </h3>
          <Button
            type="secondary"
            onClick={() => {
              onCopy();
            }}
          >
            {#if copied}
              <Check size="16px" />
              {m.avatar_copied_url()}
            {:else}
              <Link size="16px" className="text-primary-500" />
              {m.avatar_copy_url_for_view()}
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
          <div class="flex flex-col gap-y-4">
            <ExportDashboardForm
              runExport={runPdfExport}
              showTabOptions={hasTabGroups}
              onComplete={() => (isOpen = false)}
            />
            {#if $reports}
              <Button
                type="secondary"
                onClick={() => {
                  showScheduledReportDialog = true;
                  isOpen = false;
                }}
              >
                {m.report_form_schedule_pdf_button()}
              </Button>
            {/if}
          </div>
        </TabsContent>
      {/if}
    </Tabs>
  </PopoverContent>
</Popover>

{#if showScheduledReportDialog && canvasName}
  <ScheduledReportDialog
    bind:open={showScheduledReportDialog}
    props={{ mode: "create-canvas", canvasName }}
  />
{/if}

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
