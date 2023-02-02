<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";

  const dispatch = createEventDispatcher();
</script>

<div class="flex flex-col gap-y-3 my-3 px-3">
  <div class="flex flex-col gap-y-1">
    <label for="start-date" style="font-size: 10px;">Start date</label>
    <input type="datetime-local" name="start-date" />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" style="font-size: 10px;">End date</label>
    <input type="datetime-local" name="end-date" />
  </div>
  <div class="flex">
    <div class="flex-grow" />
    <Button type="primary" on:click={() => dispatch("apply")}>Apply</Button>
  </div>
</div>
