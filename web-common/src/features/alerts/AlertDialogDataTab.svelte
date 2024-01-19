<script lang="ts">
  import FormSection from "../../components/forms/FormSection.svelte";
  import InputV2 from "../../components/forms/InputV2.svelte";
  import Select from "../../components/forms/Select.svelte";
  import DataPreview from "./DataPreview.svelte";

  let title: string;
  let titleError: string;
  let splitByDimension: string;
  let forEvery: string;
</script>

<div class="flex flex-col gap-y-5">
  <FormSection title="Alert name">
    <InputV2
      bind:value={title}
      error={titleError}
      id="title"
      placeholder="My alert"
    />
  </FormSection>
  <FormSection
    title="Filters"
    description="These are inherited from the underlying dashboard view."
  ></FormSection>
  <FormSection
    title="Alert data"
    description="Select the measures you want to monitor."
  >
    <InputV2
      bind:value={title}
      error={titleError}
      id="measures"
      placeholder="Add measures"
    />
    <Select
      bind:value={splitByDimension}
      id="dimensionSplit"
      label="Split by dimension"
      options={["Dim1", "Dim2", "Dim3"].map((dimension) => ({
        value: dimension,
      }))}
    />
    <Select
      bind:value={forEvery}
      id="forEvery"
      label="For every"
      options={["Interval1", "Interval2", "Interval3"].map((timeInterval) => ({
        value: timeInterval,
      }))}
    />
  </FormSection>
  <FormSection title="Data preview">
    <DataPreview />
  </FormSection>
</div>
