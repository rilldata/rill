<script lang="ts">
  export let color: "default" | "red" | "blue" = "default";

  function getColorClass(color: string) {
    switch (color) {
      case "default":
        return "bg-gray-100 border-gray-300 text-gray-600";
      case "red":
        return "bg-red-50 border-red-300 text-red-600";
      case "blue":
        return "bg-blue-50 border-blue-300 text-blue-600";
    }
  }
</script>

<div
  class="px-2 h-[21px] border rounded-[20px] items-center justify-center inline-flex shrink-0 {getColorClass(
    color,
  )}"
>
  <span class="text-xs font-normal">
    <slot />
  </span>
</div>
