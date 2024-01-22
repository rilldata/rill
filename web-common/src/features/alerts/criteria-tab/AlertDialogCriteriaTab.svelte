<script lang="ts">
  import { page } from "$app/stores";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import DataPreview from "@rilldata/web-common/features/alerts/DataPreview.svelte";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, updateField } = formState;

  $: dashboardName = $page.params.dashboard;

  function handleCriteriaUpdate(expr: V1Expression) {
    // TODO: support multiple groups
    updateField("criteria", expr);
  }
</script>

<div class="flex flex-col gap-y-5 p-2.5 max-h-96 bg-gray-100 overflow-scroll">
  <FormSection
    description="Trigger alert when these conditions are met"
    title="Criteria"
  >
    <CriteriaGroup
      expr={$form["criteria"]}
      on:update={(e) => handleCriteriaUpdate(e.detail)}
    />
  </FormSection>
  <FormSection title="Alert Preview">
    <DataPreview
      criteria={$form["criteria"]}
      dimension={$form["splitByDimension"]}
      measure={$form["measure"]}
      metricsView={dashboardName}
    />
  </FormSection>
</div>
