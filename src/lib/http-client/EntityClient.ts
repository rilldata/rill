import { HttpStreamClient } from "$lib/http-client/HttpStreamClient";
import { config } from "$lib/application-state-stores/application-store";

export class EntityClient<Entity extends Record<string, unknown>> {
  public static instance: EntityClient<Record<string, unknown>>;

  constructor(
    private readonly entityMap: {
      [key in keyof Entity]?: { endPoint: string; field: string };
    },
    private readonly pathGetter: (
      id: string,
      ...otherIds: Array<string>
    ) => string
  ) {}

  public create(...otherIds: Array<string>) {
    return HttpStreamClient.instance.request(
      this.pathGetter(null, ...otherIds),
      "PUT"
    );
  }

  public async getAll(...otherIds: Array<string>) {
    return (
      await (
        await fetch(
          `${config.server.serverUrl}/api${this.pathGetter(null, ...otherIds)}`
        )
      ).json()
    ).data;
  }

  public async updateField(
    id: string,
    name: keyof Entity,
    value: unknown,
    ...otherIds: Array<string>
  ) {
    if (name in this.entityMap) {
      return HttpStreamClient.instance.request(
        `${this.pathGetter(id, ...otherIds)}/${this.entityMap[name].endPoint}`,
        "POST",
        { [this.entityMap[name].field]: value }
      );
    } else {
      return HttpStreamClient.instance.request(
        this.pathGetter(id, ...otherIds),
        "POST",
        { [name]: value }
      );
    }
  }

  public delete(id: string, ...otherIds: Array<string>) {
    return HttpStreamClient.instance.request(
      this.pathGetter(id, ...otherIds),
      "DELETE"
    );
  }
}
