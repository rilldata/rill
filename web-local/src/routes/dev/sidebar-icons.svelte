<script>
  import SurfaceViewIcon from "@rilldata/web-common/components/icons/SurfaceView.svelte";
  let mode = "hamburger;";
</script>

<button
  on:click={() => {
    mode = "left";
  }}
>
  left
</button>
<button
  on:click={() => {
    mode = "hamburger";
  }}
>
  hamburger
</button>

<button
  on:click={() => {
    mode = "right";
  }}
>
  right
</button>

<SurfaceViewIcon size="5rem" {mode} />
<SurfaceViewIcon size="4rem" {mode} />
<SurfaceViewIcon size="3rem" {mode} />
<SurfaceViewIcon size="2rem" {mode} />
<SurfaceViewIcon size="24px" {mode} />
<SurfaceViewIcon size="1rem" {mode} />
