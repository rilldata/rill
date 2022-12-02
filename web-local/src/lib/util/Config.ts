export type NonFunctionPropertyNames<T> = {
  [K in keyof T]: T[K] extends (...args: unknown[]) => unknown ? never : K;
}[keyof T];
export type NonFunctionProperties<T> = Pick<T, NonFunctionPropertyNames<T>>;

export type ConfigFieldType = {
  field: string;
  key?: string;
  defaultValue?: unknown;
  subConfigClass?: typeof Config;
};

/**
 * Config class that supports defaults.
 * Usage: Extend this class and pass the config class name as generic arg to add style to constructor.
 */
export class Config<C> {
  private static configTypes = new Array<ConfigFieldType>();

  constructor(_: {
    [K in keyof NonFunctionProperties<C>]?: NonFunctionProperties<C>[K];
  }) {
    // no-op
  }

  /**
   * Decorator that adds a config field.
   * Automatically pulls the field from config object passed in constructor.
   *
   * @param defaultValue Defaults to undefined. Can be overridden with this param.
   * @param key Defaults to using property key. Can be overridden with this param.
   */
  public static ConfigField(defaultValue?: unknown, key?: string) {
    return (target: Config<unknown>, propertyKey: string) => {
      const constructor = this.createConfigTypes(target);
      constructor.configTypes.push({
        field: propertyKey,
        key: key ?? propertyKey,
        defaultValue,
      });
    };
  }

  /**
   * Decorator that adds a sub config field. Takes a class also extends Config.
   * Automatically pulls the field from config object passed in constructor.
   *
   * @param subConfigClass
   * @param defaultValue Defaults to empty object. Can be overridden with this param.
   * @param key Defaults to using property key. Can be overridden with this param.
   */
  public static SubConfig(
    subConfigClass: typeof Config,
    defaultValue?: unknown,
    key?: string
  ) {
    return (target: Config<unknown>, propertyKey: string) => {
      const constructor = this.createConfigTypes(target);
      constructor.configTypes.push({
        field: propertyKey,
        key: key ?? propertyKey,
        defaultValue: defaultValue ?? {},
        subConfigClass,
      });
    };
  }

  private static createConfigTypes(target: Config<unknown>) {
    const constructor = target.constructor as typeof Config;

    if (!Object.prototype.hasOwnProperty.call(constructor, "configTypes")) {
      constructor.configTypes = [...constructor.configTypes];
    }

    return constructor;
  }

  protected setFields(configJson: {
    [K in keyof NonFunctionProperties<C>]?: NonFunctionProperties<C>[K];
  }) {
    // if null is passed explicitly default param value does not take hold
    configJson = configJson || {};
    (this.constructor as typeof Config).configTypes.forEach((configType) => {
      const configValue =
        configJson[configType.key ? configType.key : configType.field] ??
        configType.defaultValue;

      if (configType.subConfigClass) {
        this[configType.field] = new configType.subConfigClass(configValue);
      } else {
        this[configType.field] = configValue;
      }
    });
  }
}
