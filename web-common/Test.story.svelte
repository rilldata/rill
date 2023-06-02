<script lang="ts">
  export let Hst;
  const foo: string = "FOOOFOOFO";
</script>

<Hst.Story>
  <Hst.Variant>🐱</Hst.Variant>

  <Hst.Variant>
    <div class="bg-purple-400">{foo}</div>
  </Hst.Variant>
</Hst.Story>
