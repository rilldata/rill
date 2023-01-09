<script lang="ts">
  import EmbeddedSourceSet from "./EmbeddedSourceSet.svelte";
  const links = [
    "https://raw.githubusercontent.com/hhllcks/snldb/master/output/actors.csv",
    "https://raw.githubusercontent.com/hhllcks/snldb/master/output/appearances.csv",
    "https://raw.githubusercontent.com/hhllcks/snldb/master/output/casts.csv",
    "https://raw.githubusercontent.com/hhllcks/snldb/master/output/characters.csv",
    "https://raw.githubusercontent.com/hulmer/dataset/master/output/dataset2.csv",
    "https://raw.githubusercontent.com/hulmer/dataset/master/output/dataset3.csv",
    "https://raw.githubusercontent.com/hulmer/dataset/master/output/dataset1.csv",
    "s3://tvs-bidding-engine-dev/prediction/caldera_lab/visited_website/prediction_2022-02-18-12.txt",
    "s3://tvs-bidding-engine-dev/prediction/caldera_lab/visited_website/prediction_2022-10-22.txt",
    "s3://tvs-bidding-engine-dev/prediction/caldera_lab/whatever/this/is/a/long/bucket/prediction_2022-10-22.txt",
    "s3://whatever/path/to/bucket/year=[2012]/month=[10]/day=[21]/partition-000000000000.parquet",
    "gs://tvs-bidding-engine-dev/prediction/caldera_lab/whatever/prediction_2022-10-22.txt",
    "gs://moz-test-bucket-123-very-long-title-that-overflows/predictions/model/xyz/partition-00000000.parquet",
  ];

  function locatorAndRest(url) {
    const [_, locator, rest] = url.trim().split(/(gs:\/\/|s3:\/\/|https:\/\/)/);
  }

  function extractDomain(url, pathSegments = 0) {
    const [locator, rest] = locatorAndRest(url);

    const [domain, ...path] = rest.split("/");
    return `${domain}/${path.slice(0, pathSegments).join("/")}`;
  }

  function groupByUniqueParts(uris: string[]) {
    // first group by domain.
    // then figure out how much of the next part to show. we do this by marching toward
    // the end of the string, and the moment the URIs don't match, we show the rest of the string to the
    // user then show uri.slice(0, lastMatchingIndex).

    const groupedURIs = uris.reduce((obj, uri) => {
      const [_, locator, rest] = uri
        .trim()
        .split(/(gs:\/\/|s3:\/\/|https:\/\/)/);

      const components = rest.split("/");
      // domain name or bucket name.
      const domainName = components[0];
      const domain =
        components.length > 2
          ? components.slice(0, 2).join("/")
          : components[0];
      obj[domain] = obj[domain] ? [...obj[domain], uri] : [uri];
      return obj;
    }, {});

    for (let domain in groupedURIs) {
      const domainURIs = groupedURIs[domain];
      // march to the end of the string.
      const longestURI = Math.max(...domainURIs.map((uri) => uri.length));
      let endingIndex = 0;
      for (let i = 0; i < longestURI; i++) {
        // the moment the set of URIs no longer matches, we stop.
        if (new Set(domainURIs.map((uri) => uri.slice(0, i))).size !== 1) {
          endingIndex = i;
          break;
        }
      }
      const identifier = domainURIs[0].split(/(gs:\/\/|s3:\/\/|https:\/\/)/)[1];
      groupedURIs[domain] = {
        endingIndex,
        domain,
        which: identifier.replace("://", ""),
        leftPart: domainURIs[0]
          .slice(0, endingIndex - 1)
          .replace(identifier, ""),
        uris:
          domainURIs.length > 1
            ? domainURIs.map((uri) => uri.slice(endingIndex - 1))
            : domainURIs.map((uri) => uri.split("/").at(-1)),
      };
    }
    return groupedURIs;
  }

  let chunksOfLinks = groupByUniqueParts(links);
  let sections = Object.keys(chunksOfLinks).reduce((acc, v) => {
    acc[v] = true;
    return acc;
  }, {});

  function domainAndRest(uri: string) {
    const s = uri.split("/");
    return [s[0], s.slice(1).join("")];
  }
</script>

<div class="flex flex-col gap-y-4">
  <section>
    <h1>default state – not great!</h1>
    <div style:width="300px" style:outline="1px solid lightgray">
      {#each links as link}
        <div class="text-ellipsis overflow-hidden whitespace-nowrap">
          {link}
        </div>
      {/each}
    </div>
  </section>
  <section>
    <h1>proposal</h1>
    <div
      style:width="300px"
      style:outline="1px solid lightgray"
      class="py-3 space-y-2"
    >
      {#each Object.keys(chunksOfLinks) as domain, i}
        {@const domainSet = chunksOfLinks[domain]}
        {@const leftPart = domainSet.leftPart}
        {@const links = domainSet.uris}
        {@const which = domainSet.which}
        <EmbeddedSourceSet location={domain} type={which} sources={links} />
      {/each}
    </div>
  </section>
</div>
