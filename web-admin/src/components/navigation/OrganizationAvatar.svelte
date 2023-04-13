<script lang="ts">
  export let organization: string;

  function getColorClass(organization: string) {
    const hash = organization.split("").reduce((acc, char) => {
      return acc + char.charCodeAt(0);
    }, 0);
    const colors = [
      "bg-blue-700",
      "bg-green-700",
      "bg-yellow-700",
      "bg-red-700",
      "bg-purple-700",
      "bg-pink-700",
      "bg-indigo-700",
      "bg-gray-700",
    ];
    return colors[hash % colors.length];
  }
</script>

<div
  class="inline-flex items-center justify-center w-3 h-3 rounded text-white font-bold {getColorClass(
    organization
  )}"
  style="font-size: 9px; line-height: .5rem;"
>
  {organization.charAt(0).toUpperCase()}
</div>
