<script lang="ts">
  type Color =
    | "red"
    | "blue"
    | "green"
    | "purple"
    | "orange"
    | "yellow"
    | "gray";

  export let color: Color = "gray";

  function getColorClass(color: string) {
    switch (color) {
      case "red":
        return "bg-red-50 border-red-300 text-red-600";
      case "blue":
        return "bg-blue-50 border-blue-300 text-blue-600";
      case "green":
        return "bg-green-50 border-green-300 text-green-600";
      case "purple":
        return "bg-purple-50 border-purple-300 text-purple-600";
      case "orange":
        return "bg-orange-50 border-orange-300 text-orange-600";
      case "yellow":
        return "bg-yellow-50 border-yellow-300 text-yellow-600";
      case "gray":
        return "bg-gray-100 border-gray-300 text-gray-600";
      default:
        return "bg-gray-100 border-gray-300 text-gray-600";
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
