import "../moduleAlias";
import { parseExpression } from "$common/utils/parseQuery";

console.log(JSON.stringify(parseExpression("avg(a) - sum(b)"), null, 2));
console.log(JSON.stringify(parseExpression("a"), null, 2));
console.log(JSON.stringify(parseExpression("distinct a"), null, 2));
