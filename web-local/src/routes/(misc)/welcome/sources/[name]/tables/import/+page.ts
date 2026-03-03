import { ImportTableRunner } from "@rilldata/web-common/features/sources/import/ImportTableRunner.ts";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

export function load({ params, url: { searchParams } }) {
  const connectorName = params.name;
  const database = searchParams.get("database") ?? "";
  const schema = searchParams.get("schema") ?? "";
  const table = searchParams.get("table");
  const name = searchParams.get("name");
  // TODO: is there a better way to do pass yaml?
  const yaml = sessionStorage.getItem("yaml");

  if (!table || !name || !yaml) {
    console.error("Missing required parameters");
    throw redirect(302, "/");
  }

  const runner = new ImportTableRunner(
    get(runtime).instanceId,
    name,
    {
      connector: connectorName,
      database,
      schema,
      table,
    },
    yaml,
  );
  void runner.run();

  return {
    runner,
  };
}
