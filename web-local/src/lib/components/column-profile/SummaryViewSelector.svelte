<script lang="ts">
  import * as classes from "../../util/component-classes";
  import { defaultSort, sortByNullity, sortByName } from "./sort-utils";

  export let sortMethod = defaultSort;
  export let previewView = "summaries";
</script>

<div class="flex justify-between text-gray-500">
  <select bind:value={sortMethod} class={classes.NATIVE_SELECT}>
    <!-- <option value={sortByOriginalOrder}>sort by original order</option> -->
    <option value={defaultSort}>sort by data type</option>
    <option value={sortByNullity}>sort by null %</option>
    <option value={sortByName}>sort by name</option>
  </select>
  <select bind:value={previewView} class={classes.NATIVE_SELECT}>
    <option value="summaries">show summaries</option>
    <option value="example">show example</option>
  </select>
</div>
