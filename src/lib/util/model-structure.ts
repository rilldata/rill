/**
 * Below are a set of brute-force tools to solve common query inspection tasks:
 * - can I isolate and inspect the CTEs of my query?
 * - can I extract the column names?
 * 
 * These functions all assume that the query is properly formed 
 * in the first place; that is, there are no syntax errors.
 */

interface CTE {
    start:number,
    end:number,
    name:string,
    substring:string
}

const viableReferences = (name:string) => {
    return [
        `join ${name} `, 
        `join '${name}' `,
        `join "${name}" `, 
        `from ${name} `, 
        `from '${name}' `,
        `from "${name}" `,
    ]
}

/**
 * Sorts according to which CTE has the other as a dependency.
 * This is useful when we want to re-order CTEs to materialize them
 * according to dependence on existing CTEs.
 * @param a a CTE
 * @param b another CTE
 * @returns numbers (0, 1, -1)
 */
export function sortByCTEDependency(a:CTE, b:CTE) {
    const bQuery = b.substring.toLowerCase() + ' ';
    const aSet = viableReferences(a.name);
    const aReferencedInB = aSet.some((reference) => bQuery.includes(reference));
    // check to see if the a CTE alias is mentioned in b somehow
    if (aReferencedInB) { return -1 };
    const aQuery = a.substring.toLowerCase() + ' ';
    const bSet = viableReferences(b.name);
    const bReferencedInA = bSet.some((reference) => bQuery.includes(reference));
    // check to see if the b CTE alias is mentioned in a somehow
    if (bReferencedInA) { return 1 };
    // there is no inclusion.
    return 0;
}

function firstCharacterAt(string) {
   const output = string.split('').findIndex(char => !['\t', '\n', '\r', ' ', ';'].includes(char));
   return output === -1 ? 0 : output;
}

export function extractCTEs(query:string) : CTE[] {
    if (!query) { return []; }
    if (query.toLowerCase().trim().startsWith('select ')) { return [] };

    // Find the very start of of the expression.
    const withExpressionStartPoint = query.toLowerCase().indexOf('with ');

    // this is the start of the tape, where teh first CTE alias should be.
    let si = withExpressionStartPoint + 'WITH '.length;

    // exit if there is no `WITH` statement.
    if (si === -1) return undefined;
    const CTEs:CTE[] = [];
    // set the tape at the start.
    let ri = si;
    // the expression index.
    let ei
    // this will track the nested parens level.
    let nest = 0;
    let inside = false;
    let currentExpression = {} as CTE;
    let reachedEndOfCTEs = false;
    while (ri < query.length && !reachedEndOfCTEs) {
        let char = query[ri];

        // let's get the name of this thing.
        // we should only trigger this if nest === 1; otherwise we're selecting the CTEs of CTEs :)
        if (nest === 1 && query.slice(si, ri).toLowerCase().endsWith(' as (')) {
            currentExpression.name = query.slice(si, ri - 4).replace(',', '').trim();
        }

        // Let's up the nest by one if we encounter an open paren.
        // we will set the inside flag to true, then set the expression index
        // to match the right index.
        if (char === '(') {
            nest += 1;
            if (!inside) {
                inside = true;
                ei = ri;
            }
        }

        // If we encounter a close paren, let's unnest.
        if (char === ')') {
            nest -= 1;
        }

        // if we encounter a close parent AND the nest is at 0, then we've found the end of the CTE.
        if (char ===')' && nest === 0) {
            // we reset.
            currentExpression.start = ei + 1;
            // move up the start to the first non-whitespace char.

            const firstRealChar = firstCharacterAt(query.slice(currentExpression.start));
            if (firstRealChar !== -1) {
                currentExpression.start += firstRealChar;
            }
            currentExpression.end = ri;
            currentExpression.substring = query.slice(ei, ri+1).slice(1, -1).trim();
            CTEs.push({...currentExpression});
            si = ri+1;
            // reset the expression
            currentExpression = {} as CTE;
            nest = 0;
            inside = false;
        }

        ri += 1;
        // do we kill things if SELECT is at the end?
        if (!inside && query.slice(si, ri).trim().toLowerCase().startsWith('select ')) {
            reachedEndOfCTEs = true;
        }
    }
    return CTEs;
}

export function extractCoreSelectStatements(query:string) {
    const ctes = extractCTEs(query);
    const latest = ctes.slice(-1)[0].end;
    const restOfQuery = query.slice(latest + 1).replace(/[\s\n\t\r]/g, ' ');
    const startingBuffer = firstCharacterAt(restOfQuery);

    if (!restOfQuery.toLowerCase().trim().startsWith('select ')) {
        throw Error(`rest of query must start with select, instead with ${restOfQuery.slice(0,10)}`);
    }
    let i = 'SELECT '.length + startingBuffer;
    let ri = i;
    let ei = ri;
    let reachedFrom = false;
    let nestLevel = 0;
    let columnSelects = [];
    
    function queryEndsWithFrom(query) {
        return query.toLowerCase().endsWith('from ');
    }
    // goal:
    // when you hit AS or a , at nest level 0 (or reach nest 0 FROM statement), 
    // then put that as a column: {expression: ___, name: ____}
    while (!reachedFrom && ri < restOfQuery.length) {
        ri += 1;
        // start with the valid query situation. 
        // We split on " as " or " AS " 
        // and then make the expression the first part
        // and then the name the second
        // if no `as` then the expression is the name
        // later, we crawl the expression for the column names.
        let querySoFar = restOfQuery.slice(ei, ri).toLowerCase()
        let endsWithFrom = queryEndsWithFrom(querySoFar)
             && 
            nestLevel === 0;
        if ((restOfQuery[ri] === ',' && nestLevel === 0) || endsWithFrom) {
            
            let start = ei;
            const firstRealChar = firstCharacterAt(restOfQuery.slice(start));
            if (firstRealChar !== -1) {
                start += firstRealChar;
            }
            let end = ri;
            if (endsWithFrom) {
                end -= ' from '.length;
                // remove ` from `
            }
            // we need to set ri off right.
            //const end = ri - (endsWithFrom ? ' from '.length : 0);

            const columnExpression = restOfQuery.slice(start, end);

            const hasAs = columnExpression.toLowerCase().indexOf(' as ');
            let name;
            let expression;
            if (hasAs !== -1) {
                expression = columnExpression.slice(0, hasAs);
                name = columnExpression.slice(hasAs + 4); // remove the comma
            } else {
                expression = columnExpression;
                name = expression;
            }
            
            // look at end of name and trim the end.

            function getStartAndEnd(string) {
                let start = firstCharacterAt(string);
                let end = firstCharacterAt(string.split('').reverse().join(''));
                return [start, end]
            }
            let [_, nameEnd] = getStartAndEnd(name);
            let [expressionStart, __] = getStartAndEnd(expression);
            // look at the start of expression and trim the start.


            columnSelects.push({ 
                name: name.trim(), 
                expression: expression.trim(), 
                start: latest + start + 1 + expressionStart, 
                end: latest  + end + 1 - nameEnd});
            // move ri past the comma
            ri += 1;
            ei = ri;
        } else if (restOfQuery[ri] === '(') {
            nestLevel += 1;
        } else if (restOfQuery[ri] === ')') {
            nestLevel -= 1;
        }
    
        // cut the tape
        if (endsWithFrom) {
            reachedFrom = true;
        }
    };
    return columnSelects;
}

const expressionTokens = [
    ' where ', ' group by ', ' having ', ' order by ',
    ' left outer join ', ' right outer join ', 
    ' inner join ', ' outer join ',
    ' natural join ', ' right join ', ' left join ',
    ' limit ',
    ' join ', // make join last,
]

function getAllIndexes(arr, val) {
    const indexes = []
    let i = -1;
    while ((i = arr.indexOf(val, i+1)) != -1){
        indexes.push(i);
    }
    return indexes;
}

export function extractFromStatements(query:string) {

    let latest = 0;
    let restOfQuery = query.replace(/[\s\t\r\n]/g, ' ');
    // replace -- comments here?
    restOfQuery = restOfQuery.replace(/-- .*\n*/g, ' ')
    const finds = getAllIndexes(restOfQuery.toLowerCase(), ' from ');
    
    let sourceTables = [];

    finds.forEach((fi) => {
        let ei = fi + ' from '.length;
        let ri = ei;
        let nestLevel = 0;
        while (ri < restOfQuery.length) {
            
            ri +=1;
            let char = restOfQuery[ri-1];
            let seqSoFar = restOfQuery.slice(ei, ri);
            
            // skip if the FROM statement contains a nested statement inside it.
            if (char === '(' && nestLevel === 0) {
                break;
            }
            
            const containsExpressionToken = (str) => {
                return expressionTokens.some(token => str.endsWith(token));
            }
            if (containsExpressionToken(seqSoFar.toLowerCase()) || char === ';' || char === ')' || ri === restOfQuery.length) {

                // reset seqSoFar to not include th expression token.

                expressionTokens.forEach((token) => {
                    // remove the token?
                    if (seqSoFar.toLowerCase().endsWith(token)) {
                        seqSoFar = seqSoFar.slice(0, -token.length);
                        ri = ri - token.length;
                    }
                })

                // we hit the end of the table def.
                let rightCorrection = 0;
                let leftCorrection = 0;

                // can we flip this?
                // get the right side if there's extra characters;
                const leftSide = firstCharacterAt(seqSoFar);
                const rightSide = firstCharacterAt(seqSoFar.split('').reverse().join(""));

                if (rightSide !== -1) {
                    rightCorrection = rightSide;
                }
                if (leftSide !== -1) {
                    leftCorrection = leftSide;
                }

                if (char === ")") {
                    rightCorrection += 1;
                    // look for spaces b/t ) and statement.
                    let additionalRightBuffer = firstCharacterAt(seqSoFar.split('').reverse().join("").slice(1));
                    if (additionalRightBuffer !== -1) {
                        rightCorrection += additionalRightBuffer;
                    }
                }
                
                const finalSeq = restOfQuery.slice(ei + leftCorrection, ri - rightCorrection);
                
                sourceTables.push({
                    name: finalSeq,
                    start: ei  + latest + leftCorrection,
                    end: ri - rightCorrection + latest
                });

                break;
            }
        }
    })
    return sourceTables;
    // get all FROM locations.
}

const postWhereClauses = [
    ' having ', ' group by ', ' order by ', 
]

function endsWith(string, clauses = postWhereClauses) {
    let intermediate = string.toLowerCase();
    return clauses.some(clause => intermediate.endsWith(clause));
}

export function extractCoreWhereClauses(query:string) {
    // set aside CTEs.
    const ctes = extractCTEs(query);
    let latest = 0;
    let restOfQuery = query;
    if (ctes.length) {
        latest = ctes.slice(-1)[0].end;
        restOfQuery = query.slice(latest + 1);
    }
    const startingBuffer = firstCharacterAt(restOfQuery);

    if (!restOfQuery.toLowerCase().trim().startsWith('select ')) {
        throw Error(`rest of query must start with select, instead with ${restOfQuery.slice(0,10)}`);
    }
    let i = 'SELECT '.length + (startingBuffer !== -1 ? startingBuffer : 0);
    let ri = i;
    let ei = ri;
    let inWhereClause = false;
    let nestLevel = 0;

    const whereClauses = [];

    while (ri < restOfQuery.length) {
        // match on WHERE.
        ri += 1;
        const substring = restOfQuery.slice(ei, ri);
        const normalized = substring.replace(/[\t\r\n\s]/g, ' ').toLowerCase();
        if (normalized.endsWith(' where ')) {
            // set ei to ri;
            inWhereClause = true;
            // reset ei to equal ri.
            ei = ri;
        }
        if (inWhereClause) {
            console.log('in the where clause!')

            // we should actually just go nuts here.
            if (ri === restOfQuery.length || substring[ri] === ';' || endsWith(substring, postWhereClauses) && nestLevel === 0) {
                // let's split on  ' and ' and ' or '
                const whereClause = substring; // we need to massage this
                let matches = [];
                const andsAndOrs = [...getAllIndexes(normalized, ' and '), getAllIndexes(normalized, ' or ')];
            }

            // wait until we get to an AND or OR statement.
            if ((ri === restOfQuery.length || normalized.endsWith(' and ') || normalized.endsWith(' or ')) && nestLevel === 0) {
                const ci = ei;
                // this is where we split up and say the relation.
                // ok let's goooo
                let clauseType = normalized.endsWith(' and ') ? ' and ' : ' or ';
                let start = latest + ei + firstCharacterAt(normalized);
                // append the clause
                let statement = substring.slice(0, -clauseType.length);
                let end = ri + latest - firstCharacterAt(statement.split('').reverse().join(''));
                whereClauses.push({
                    start, 
                    end, 
                    clauseType: clauseType.trim(),
                    statement
                })
                // rest ei to ri and continue the tape.
                ei = ri;

            }
            // check for nesting of subqueries and also function parens.

            if (restOfQuery[ri] === '(') {
                nestLevel += 1;
            }
            if (restOfQuery[ri] === ')') {
                nestLevel -= 1;
            }
        }
        // if the substring ends in an expression token that takes us out of the where statement,
        // and the nest level is 0, let's abort.
        if ((endsWith(substring, postWhereClauses) && nestLevel === 0) || (nestLevel === 0 && substring[ri] === ';')) {
            break;
        }
    }
    return whereClauses;
}

export function extractJoins(query) {
    let latest = 0;
    let restOfQuery = query.replace(/[\s\t\r\n]/gi, ' ');

    const finds = getAllIndexes(restOfQuery.toLowerCase(), ' join ');
    let matches = [];
    for (let find of finds) {
        let ei = find + ' join '.length;//restOfQuery.slice(find + ' join '.length);
        let ri = ei;
        let noMatchYet = true;
        while (ri < restOfQuery.length || noMatchYet) {
            ri += 1;
            let soFar = restOfQuery.slice(ei, ri);
            // break if we're on a subquery, for now.
            if (restOfQuery[ri] === '(') {
                break;
            }
            if (soFar.toLowerCase().endsWith(' on ')) {
                // we have a match.
                let name = soFar.slice(0, -` on `.length);
                let start = ei + firstCharacterAt(name);
                let end = ei + name.length - firstCharacterAt(name.split('').reverse().join(''));
                name = name.trim();
                matches.push({name, start, end});
                break;

            }
        }
        
    }
    return matches;
}

export function extractSourceTables(query) {
    const CTEs = extractCTEs(query);
    let cteAliases = new Set(CTEs.map(cte => cte.name));
    const froms = extractFromStatements(query).filter(statement => !cteAliases.has(statement.name));
    const joins = extractJoins(query).filter(statement => !cteAliases.has(statement.name));
    const all = [...froms.map(table => ({...table, type: 'from'})), ...joins.map(table => ({...table, type: 'join'}))];
    // now let's coalesce.
    // get uniques.
    const names = [...new Set(all.map(table => table.name))];
    return names.map((name) => {
        const tables = all.filter(table => table.name === name);
        return {
            name,
            tables
        }
    })
}