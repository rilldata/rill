<script lang="ts">
  let theme = localStorage.getItem("theme") || "light";
  function toggleTheme() {
    theme = theme === "light" ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", theme);
  }
</script>

<button on:click={toggleTheme}> {theme} </button>
