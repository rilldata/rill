import { fetchResources } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { writable, type Writable } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export class Resource {}

type FilePath = string;

export class Resources {
  private readonly resources: Map<FilePath, Writable<V1Resource[]>> = new Map();

  async init(instanceId: string) {
    const allResources = await fetchResources(instanceId);

    for (const resource of allResources) {
      // set query data for GetResource to avoid refetching data we already have
      queryClient.setQueryData(
        getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.name": resource.meta?.name?.name,
          "name.kind": resource.meta?.name?.kind,
        }),
        {
          resource,
        },
      );

      const meta = resource.meta;
      if (!meta) continue;

      const path = meta?.filePaths?.pop();
      if (!path) continue;

      const name = meta.name;
      if (!name) continue;

      const paths = this.resources.get(path);

      if (!paths) {
        this.resources.set(path, writable([resource]));
      } else {
        paths.update((resources) => {
          return [...resources, resource];
        });
      }
    }
  }

  update(resource: V1Resource) {
    const meta = resource.meta;
    if (!meta) return;

    const path = meta.filePaths?.pop();
    if (!path) return;

    const paths = this.resources.get(path);

    if (!paths) {
      this.resources.set(path, writable([resource]));
    } else {
      paths.update((resources) => {
        return [...resources, resource];
      });
    }
  }

  get(path: FilePath): Writable<V1Resource[]> {
    let paths = this.resources.get(path);

    if (!paths) {
      paths = writable([]);
      this.resources.set(path, paths);
    }

    return paths;
  }
}

export const resources = new Resources();
