<p class="text-gray-300 my-1 w-full line-clamp-15">
  <slot />
</p>

<style>
  .line-clamp-15 {
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 15;
    line-clamp: 15;
    -webkit-box-orient: vertical;
  }
</style>
