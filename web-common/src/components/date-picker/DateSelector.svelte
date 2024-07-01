<script lang="ts">
  import * as Select from "@rilldata/web-common/components/select";
  import { DateTime } from "luxon";

  export let value: DateTime | undefined;
  export let maxYear: number;
  export let minYear: number;
  export let zone: string;
  export let selecting: boolean;
  export let label: string;

  $: months = Array.from({ length: 12 }, (_, i) => ({
    value: i + 1,
    label: DateTime.fromObject({ year: maxYear, month: i + 1 }).monthLong ?? "",
  }));

  $: days = value
    ? Array.from({ length: value.daysInMonth ?? 0 }, (_, i) => ({
        value: i + 1,
        label: (i + 1).toString(),
      }))
    : [];

  $: years = Array.from({ length: maxYear - minYear + 20 }, (_, i) => ({
    value: DateTime.now().year - i,
    label: (DateTime.now().year - i).toString(),
  }));
</script>

<div class="flex gap-2" class:selecting>
  <Select.Root
    onSelectedChange={(e) => {
      if (e === undefined) return;
      if (!value) {
        value = DateTime.fromObject({ year: maxYear, month: e.value }).setZone(
          zone,
        );
      } else {
        value = value.set({ month: e.value });
      }
    }}
    items={months}
  >
    <Select.Trigger class="flex-none w-32" aria-label="Select a {label} month">
      <Select.Value placeholder={value?.monthLong ?? ""} />
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each months as { value, label }}
        <Select.Item {value}>
          {label}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>

  <Select.Root
    onSelectedChange={(e) => {
      if (e === undefined) return;
      if (!value) {
        value = DateTime.fromObject({
          year: maxYear,
          month: 1,
          day: e.value,
        }).setZone(zone);
      } else {
        value = value.set({ day: e.value });
      }
    }}
    disabled={!value}
    items={days}
  >
    <Select.Trigger class="w-16" aria-label="Select a {label} day">
      <Select.Value placeholder={value?.day.toString() ?? ""} />
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each days as { value, label }}
        <Select.Item {value}>
          {label}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>

  <Select.Root
    onSelectedChange={(e) => {
      if (e === undefined) return;
      if (!value) {
        value = DateTime.fromObject({
          year: e.value,
          month: 1,
          day: 1,
        }).setZone(zone);
      } else {
        value = value.set({ year: e.value });
      }
    }}
    items={years}
  >
    <Select.Trigger class="w-24" aria-label="Select a {label} year">
      <Select.Value placeholder={value?.year.toString() ?? ""} />
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each years as { value, label }}
        <Select.Item {value}>
          {label}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
</div>

<style lang="postcss">
  :global(.selecting button) {
    @apply ring-1 ring-primary-400;
  }
</style>
