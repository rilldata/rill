<header>
  <div
    class="grid justify-between pl-3 pr-1 items-center"
    style:grid-template-columns="max-content max-content"
  >
    <div class="font-bold">
      <slot name="title" />
    </div>
    <div>
      <slot name="right" />
    </div>
  </div>
</header>
