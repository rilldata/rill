<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    DEFAULT_LABELS_FORMAT,
    DEFAULT_LABELS_THRESHOLD,
    type LabelsConfig,
    type LabelsFormat,
  } from "@rilldata/web-common/features/components/charts/circular/constants";

  export let key: string;
  export let label: string;
  export let value: LabelsConfig | undefined;
  export let onChange: (next: LabelsConfig | undefined) => void;

  $: show = value?.show === true;

  const formatOptions: { value: LabelsFormat; label: string }[] = [
    { value: "percent", label: "Percent" },
    { value: "value", label: "Value" },
  ];

  function toggleShow() {
    if (show) {
      onChange(undefined);
    } else {
      onChange({
        show: true,
        format: DEFAULT_LABELS_FORMAT,
        threshold: DEFAULT_LABELS_THRESHOLD,
      });
    }
  }

  function setFormat(next: string) {
    onChange({
      ...(value ?? {}),
      show: true,
      format: next as LabelsFormat,
    });
  }

  function setThreshold(next: number | undefined) {
    onChange({
      ...(value ?? {}),
      show: true,
      threshold: next,
    });
  }
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex justify-between py-1 items-center">
    <InputLabel small {label} id={key} faint={!show} />
    <Switch checked={show} onclick={toggleShow} small />
  </div>

  {#if show}
    <div class="flex flex-col gap-y-2 pl-2">
      <div class="flex justify-between items-center gap-x-2">
        <InputLabel small label="Format" id="{key}-format" />
        <div class="control">
          <Select
            id="{key}-format"
            label=""
            size="sm"
            full={true}
            sameWidth
            fontSize={12}
            options={formatOptions}
            value={value?.format ?? DEFAULT_LABELS_FORMAT}
            onChange={setFormat}
          />
        </div>
      </div>
      <div class="flex justify-between items-center gap-x-2">
        <InputLabel small label="Hide labels below (%)" id="{key}-threshold" />
        <div class="control">
          <Input
            id="{key}-threshold"
            label=""
            inputType="number"
            size="sm"
            value={value?.threshold ?? ""}
            oninput={(e: Event) => {
              const target = e.currentTarget;
              if (!(target instanceof HTMLInputElement)) return;
              if (target.value === "") {
                setThreshold(undefined);
              } else {
                const n = target.valueAsNumber;
                setThreshold(Number.isNaN(n) ? undefined : n);
              }
            }}
          />
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .control {
    @apply w-32 shrink-0;
  }
</style>
