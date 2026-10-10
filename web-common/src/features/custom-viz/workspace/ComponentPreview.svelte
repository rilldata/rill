<script lang="ts">
  import ResolvedComponentChart from "@rilldata/web-common/features/custom-viz/ResolvedComponentChart.svelte";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { sendComponentFilePrompt } from "@rilldata/web-common/features/custom-viz/component-ai-agent";
  import { boundMetricsViewName } from "@rilldata/web-common/features/custom-viz/params";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    V1TimeGrain,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let componentName: string;
  export let resource: V1Resource | undefined;
  export let args: Record<string, unknown>;
  // The component's file path; enables the "fix with AI" action.
  export let filePath: string | undefined = undefined;

  const client = useRuntimeClient();

  const { developerChat } = featureFlags;

  function fixWithAI(error: string) {
    if (!filePath) return;
    sendComponentFilePrompt(
      client,
      filePath,
      `It fails to render with this error: "${error}". Fix it so it renders cleanly, keeping the chart it draws, its bound channels, and whichever spec format it already uses.`,
    );
  }

  // The preview has no canvas parent, so there is no theme object to apply; only
  // the light/dark mode, which the chart's colors resolve from CSS variables.
  let themeMode: "light" | "dark";
  $: themeMode = $themeControl === "dark" ? "dark" : "light";

  $: componentSpec =
    resource?.component?.state?.validSpec ?? resource?.component?.spec;
</script>

<div class="size-full min-h-[400px] flex flex-col p-4">
  <!-- overflow-hidden bounds charts with fixed/step sizes to the preview area. -->
  <div class="flex-1 min-h-[320px] max-h-full overflow-hidden">
    <!-- Server-side resolution is only possible once the component has a valid spec;
         drafts fall back to client-side substitution.
         The preview has no dashboard time controls to inherit a grain from, so it buckets by day:
         enough to keep a time series readable, where the raw timestamps would render one mark per event. -->
    <ResolvedComponentChart
      {componentName}
      {componentSpec}
      resolvable={!!resource?.component?.state?.validSpec}
      {args}
      stateUpdatedOn={resource?.meta?.stateUpdatedOn}
      missingError={m.component_preview_no_spec()}
      metricsViewName={boundMetricsViewName(componentSpec?.params ?? [], args)}
      name={componentName}
      timeGrain={V1TimeGrain.TIME_GRAIN_DAY}
      showDataTable
      {themeMode}
    >
      <div
        slot="error"
        let:error
        class="size-full flex flex-col items-center justify-center gap-y-2"
      >
        <ComponentError {error} />
        {#if $developerChat && filePath}
          <button
            class="text-xs font-medium text-primary-600 hover:text-primary-700"
            onclick={() => fixWithAI(error)}
          >
            {m.component_fix_with_ai()}
          </button>
        {/if}
      </div>
    </ResolvedComponentChart>
  </div>
</div>
