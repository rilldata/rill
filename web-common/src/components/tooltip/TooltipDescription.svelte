<p class="text-gray-300 my-1 w-full line-clamp-3">
  <slot />
</p>

<style>
  .line-clamp-3 {
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
  }
</style>
