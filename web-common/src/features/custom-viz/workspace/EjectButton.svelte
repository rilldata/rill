<script lang="ts">
  import AlertDialogConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-confirmation.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ejectComponent } from "@rilldata/web-common/features/custom-viz/flint/eject-component";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { parseDocument, Scalar } from "yaml";

  export let fileArtifact: FileArtifact;
  export let componentName: string;
  export let resource: V1Resource | undefined;
  /** The preview's test bindings, which the spec is compiled against. */
  export let args: Record<string, unknown>;

  const client = useRuntimeClient();

  let confirming = false;
  let ejecting = false;

  // Deliberately the valid spec rather than the draft: ejecting compiles what the server resolves,
  // which is the last reconciled version. A component that already carries a Vega-Lite spec is the
  // result of ejecting and has nothing left to eject.
  $: validSpec = resource?.component?.state?.validSpec;
  $: canEject =
    validSpec?.renderer === "custom_chart" &&
    !!(validSpec?.rendererProperties as Record<string, unknown> | undefined)
      ?.spec;

  async function eject() {
    if (!resource || ejecting) return;
    confirming = false;
    ejecting = true;
    try {
      const vegaSpec = await ejectComponent(
        client,
        resource,
        componentName,
        args,
      );
      writeVegaSpecToYAML(vegaSpec);
      eventBus.emit("notification", { message: m.component_eject_success() });
    } catch (e) {
      eventBus.emit("notification", {
        message: m.component_eject_failed({
          error: e instanceof Error ? e.message : String(e),
        }),
        type: "error",
      });
    } finally {
      ejecting = false;
    }
  }

  /**
   * Swaps `custom_chart.spec` for `custom_chart.vega_spec`, leaving the params and the
   * Metrics SQL query untouched: the component's contract with dashboards does not change,
   * only how its chart is described.
   */
  function writeVegaSpecToYAML(vegaSpec: string) {
    const parsed = parseDocument(get(fileArtifact.editorContent) ?? "");
    const scalar = new Scalar(vegaSpec);
    // A block literal keeps the JSON readable and editable in the YAML file.
    scalar.type = Scalar.BLOCK_LITERAL;
    parsed.setIn(["custom_chart", "vega_spec"], scalar);
    parsed.deleteIn(["custom_chart", "spec"]);
    fileArtifact.updateEditorContent(parsed.toString(), false, true);
  }
</script>

{#if canEject}
  <Tooltip distance={8}>
    <Button
      type="secondary"
      loading={ejecting}
      loadingCopy={m.component_eject_in_progress()}
      disabled={ejecting}
      onClick={() => (confirming = true)}
    >
      <Export size="14px" />
      {m.component_eject()}
    </Button>
    <TooltipContent slot="tooltip-content" maxWidth="300px">
      {m.component_eject_tooltip()}
    </TooltipContent>
  </Tooltip>

  <AlertDialogConfirmation
    open={confirming}
    onOpenChange={(open) => (confirming = open)}
    title={m.component_eject_confirm_title()}
    description={m.component_eject_confirm_description()}
    confirmLabel={m.component_eject_confirm_action()}
    onConfirm={eject}
  />
{/if}
