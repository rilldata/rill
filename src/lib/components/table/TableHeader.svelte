<script lang="ts">
  export let position = "top";
</script>

<th
  class="
          p-0
          sticky
          {position === 'top' && 'top-0'}
          {position === 'left' && 'left-0'}
          {position === 'top-left' && 'top-0 left-0 z-10'}
      "
>
  <div
    class="
            px-4
            border
            border-gray-200
            border-t-0
            border-l-0
            bg-gray-100
            {position === 'top' && 'py-3 text-left'}
            {position === 'left' && 'py-2'}
            {position === 'top-left' && 'py-3'}
        "
  >
    <slot />
  </div>
</th>
