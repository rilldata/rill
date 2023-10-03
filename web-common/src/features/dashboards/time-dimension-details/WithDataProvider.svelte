<!--
@component
Takes in data from the runtime and manipulates it for the Time Dimension Detail table
-->
<script lang="ts">
  export let runtimeData;

  /** provide a bind site for the output */
  export let data;


  $; rowCount = 

  $: data = runtimeData;
</script>

<slot {data} />
