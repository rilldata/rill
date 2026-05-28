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
    type="single"
    value={value?.month?.toString() ?? ""}
    onValueChange={(val) => {
      if (val === undefined) return;
      const monthNum = Number(val);
      if (!value) {
        value = DateTime.fromObject({ year: maxYear, month: monthNum }).setZone(
          zone,
        );
      } else {
        value = value.set({ month: monthNum });
      }
    }}
    items={months.map((m) => ({ value: m.value.toString(), label: m.label }))}
  >
    <Select.Trigger class="flex-none w-32" aria-label="Select a {label} month">
      <span>{value?.monthLong ?? ""}</span>
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each months as { value: monthVal, label: monthLabel }}
        <Select.Item value={monthVal.toString()}>
          {monthLabel}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>

  <Select.Root
    type="single"
    value={value?.day?.toString() ?? ""}
    onValueChange={(val) => {
      if (val === undefined) return;
      const dayNum = Number(val);
      if (!value) {
        value = DateTime.fromObject({
          year: maxYear,
          month: 1,
          day: dayNum,
        }).setZone(zone);
      } else {
        value = value.set({ day: dayNum });
      }
    }}
    disabled={!value}
    items={days.map((d) => ({ value: d.value.toString(), label: d.label }))}
  >
    <Select.Trigger class="w-16" aria-label="Select a {label} day">
      <span>{value?.day.toString() ?? ""}</span>
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each days as { value: dayVal, label: dayLabel }}
        <Select.Item value={dayVal.toString()}>
          {dayLabel}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>

  <Select.Root
    type="single"
    value={value?.year?.toString() ?? ""}
    onValueChange={(val) => {
      if (val === undefined) return;
      const yearNum = Number(val);
      if (!value) {
        value = DateTime.fromObject({
          year: yearNum,
          month: 1,
          day: 1,
        }).setZone(zone);
      } else {
        value = value.set({ year: yearNum });
      }
    }}
    items={years.map((y) => ({ value: y.value.toString(), label: y.label }))}
  >
    <Select.Trigger class="w-24" aria-label="Select a {label} year">
      <span>{value?.year.toString() ?? ""}</span>
    </Select.Trigger>
    <Select.Content class="max-h-64 overflow-y-auto">
      {#each years as { value: yearVal, label: yearLabel }}
        <Select.Item value={yearVal.toString()}>
          {yearLabel}
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
