export type NonFunctionPropertyNames<T> =
    { [K in keyof T]: T[K] extends (...args: any[]) => any ? never : K }[keyof T];
export type NonFunctionProperties<T> = Pick<T, NonFunctionPropertyNames<T>>;

export type ConfigFieldType = {
    field: string;
    key?: string;
    defaultValue?: any;
    subConfigClass?: typeof Config;
}

export class Config<C> {
    private static configTypes = new Array<ConfigFieldType>();

    constructor(configJson: {
        [K in keyof NonFunctionProperties<C>]?: any
    }) {
        // if null is passed explicitly default param value does not take hold
        configJson = configJson || {};
        (this.constructor as typeof Config).configTypes.forEach((configType) => {
            const configValue = configJson[configType.key ? configType.key : configType.field] || configType.defaultValue;

            if (configValue === undefined) {
                return;
            }

            if (configType.subConfigClass) {
                this[configType.field] = new configType.subConfigClass(configValue);
            } else {
                this[configType.field] = configValue;
            }
        });
    }

    public static ConfigField(defaultValue?: any, key?: string) {
        return (target: Config<any>, propertyKey: string) => {
            const constructor = this.createConfigTypes(target);
            constructor.configTypes.push({
                field: propertyKey,
                key: key || propertyKey,
                defaultValue,
            });
        };
    }

    public static SubConfig(subConfigClass: typeof Config, defaultValue?: any, key?: string) {
        return (target: Config<any>, propertyKey: string) => {
            const constructor = this.createConfigTypes(target);
            constructor.configTypes.push({
                field: propertyKey,
                key: key || propertyKey,
                defaultValue,
                subConfigClass,
            });
        };
    }

    private static createConfigTypes(target: any) {
        const constructor: typeof Config = target.constructor;

        if (!Object.prototype.hasOwnProperty.call(constructor, "configTypes")) {
            constructor.configTypes = [...constructor.configTypes];
        }

        return constructor;
    }
}
