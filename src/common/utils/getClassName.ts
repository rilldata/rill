const CLASS_NAME_REGEX = /class (\w*) .*/;
const FUNCTION_NAME_REGEX = /function (\w*)\(.*\)/;

export function getClassName(clazz) {
    const classNameMatch = clazz.toString().match(CLASS_NAME_REGEX);
    const functionNameMatch = clazz.toString().match(FUNCTION_NAME_REGEX);
    return (classNameMatch && classNameMatch[1]) ||
        (functionNameMatch && functionNameMatch[1]);
}
