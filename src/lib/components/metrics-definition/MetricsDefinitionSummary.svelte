<script lang="ts">
  export let indentLevel: number;
  export let containerWidth: number;
</script>

<div
  class="pl-{indentLevel === 1
    ? '10'
    : '4'} pr-5 pb-2 flex justify-between text-gray-500"
  class:flex-col={containerWidth < 325}
>
  <em>((summary placeholder))</em>
</div>
