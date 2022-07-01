<script lang="ts">
  export let sticky = "top";
  export let maxWidth = "";
</script>

<th
  class="p-0"
  class:sticky={!!sticky}
  class:top-0={sticky === "top"}
  class:left-0={sticky === "left"}
  class:z-10={!!sticky}
  style={maxWidth && `max-width: ${maxWidth}`}
>
  <div
    class="
            py-3
            px-4
            border
            border-gray-200
            border-t-0
            border-l-0
            text-left
            bg-gray-100
        "
  >
    <slot />
  </div>
</th>
