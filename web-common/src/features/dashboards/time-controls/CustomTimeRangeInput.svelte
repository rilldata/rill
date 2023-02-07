<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  function applyCustomTimeRange() {
    dispatch("apply", { startDate: start, endDate: end });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      applyCustomTimeRange();
    }
  }
</script>

<div class="flex flex-col gap-y-3 my-3 px-3">
  <div class="flex flex-col gap-y-1">
    <label for="start-date" style="font-size: 10px;">Start date</label>
    <input
      bind:value={start}
      type="datetime-local"
      name="start-date"
      on:keydown={handleKeydown}
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" style="font-size: 10px;">End date</label>
    <input
      bind:value={end}
      type="datetime-local"
      name="end-date"
      on:keydown={handleKeydown}
    />
  </div>
  <div class="flex">
    <div class="flex-grow" />
    <Button type="primary" on:click={applyCustomTimeRange}>Apply</Button>
  </div>
</div>
