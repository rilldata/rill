<script lang="ts">
  import Select from "./Select.svelte";
  import { DateTime } from "luxon";

  export let value: string;
  export let id: string;
  export let label: string;

  $: currentMinute = DateTime.now().toMillis() / 1000 / 60;

  $: next15MinuteInterval = Math.ceil(currentMinute / 15) * 15;

  $: timeOptions = Array.from({ length: 24 * 4 }).map((_, index) => {
    const minutes = (next15MinuteInterval + index * 15) * 60 * 1000;

    return {
      value: DateTime.fromMillis(minutes).toFormat("HH:mm"),
      label: DateTime.fromMillis(minutes).toFormat("h:mm a"),
    };
  });
</script>

<Select {id} {label} bind:value options={timeOptions} />
