<script lang="ts">
  import { page } from "$app/stores";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { translateFilter } from "@rilldata/web-common/features/alerts/alert-filter-utils";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import DataPreview from "@rilldata/web-common/features/alerts/DataPreview.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form } = formState;

  $: dashboardName = $page.params.dashboard;
</script>

<div class="flex flex-col gap-y-5 p-2.5 max-h-96 bg-gray-100 overflow-scroll">
  <FormSection
    description="Trigger alert when these conditions are met"
    title="Criteria"
  >
    <CriteriaGroup {formState} />
  </FormSection>
  <FormSection title="Alert Preview">
    <DataPreview
      criteria={translateFilter($form["criteria"], true)}
      dimension={$form["splitByDimension"]}
      measure={$form["measure"]}
      metricsView={dashboardName}
    />
  </FormSection>
</div>
