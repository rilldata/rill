<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";

  export let key: string;
  export let label: string;
  export let options: Array<{ label: string; value: string }>;
  export let value: string;
  export let onChange: (updatedSparkline: string) => void;

  $: selected = options.findIndex((option) => option.value === value);

  function handleClick(_: number, fieldLabel: string) {
    const option = options.find((opt) => opt.label === fieldLabel);
    if (option) {
      onChange(option.value);
    }
  }
</script>

<div class="flex flex-col gap-y-2">
  <InputLabel small {label} id={`${key}-selector`} />
  <FieldSwitcher
    small
    expand
    fields={options.map((option) => option.label)}
    {selected}
    onClick={handleClick}
  />
</div>
