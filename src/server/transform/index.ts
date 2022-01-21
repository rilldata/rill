/**
 * dataset.ts
 * contains the actions that can be taken to construct a dataset.
 */

 import type { DataModelerState, Query, Item } from "src/types"
 import { sanitizeQuery as _sanitizeQuery } from "../../util/sanitize-query.js";
 import { guidGenerator } from "../../util/guid.js";

interface NewQueryArguments { 
    query?: string;
    name?: string;
    at?: number;
}

let queryNumber = 1;

export function newQuery(params:NewQueryArguments = {}): Query {
    const query = params.query || '';
    const sanitizedQuery = _sanitizeQuery(query);
    const name = `${params.name || `query_${queryNumber}`}.sql`;
    queryNumber += 1;
    return {
		query,
        sanitizedQuery,
		name,
		id: guidGenerator(),
        preview: undefined,
        sizeInBytes: undefined
	};
}

export function emptyQuery(): Query {
	return newQuery({});
}

 // TODO: we use this in other modules. Probably should have single source
 export function getByID(items:(Item[]), id:string) : Item| null {
     return items.find(q => q.id === id);
 }
 
 /**
  * NOTE: there's some amount of duplication within many of the summarizing functions.
  */
 export function createTransformActions(api) {
 
     return {
        addQuery(params:NewQueryArguments) {
            const query = params.query || undefined;
            const name = params.name || undefined;
            const at = params.at;
            return (draft:DataModelerState) => {
                if (at !== undefined) {
                    draft.queries = [...draft.queries.slice(0, at), newQuery({ query, name }), ...draft.queries.slice(at)];
                } else {
                    draft.queries.push(newQuery({ query, name })); 
                }
            };
        },
        updateQuery({id, query}) {
            return (draft:DataModelerState) => {
                const queryItem = getByID(draft.queries, id) as Query;
                queryItem.query = query;
            };
        },

        setActiveQuery({id}) {
            return (draft:DataModelerState) => {
                draft.activeQuery = id;
            }
        },

        changeQueryName({id, name}) {
            return (draft:DataModelerState) => {
                draft.queries.find((q) => q.id === id).name = name;
            }
        },

        releaseActiveQueryFocus({ id }) {
            return (draft:DataModelerState) => {
                if (draft.activeQuery === id) {
                    draft.activeQuery = undefined;
                }
            }
        },

        deleteQuery({id}) {
            return (draft:DataModelerState) => {
                draft.queries = draft.queries.filter(q => q.id !== id);
            }
        },

        moveQueryDown({id}) { 
            return (draft:DataModelerState) => {
                const idx = draft.queries.findIndex((q) => q.id === id);
                if (idx < draft.queries.length - 1) {
                    const thisQuery = { ...draft.queries[idx] };
                    const nextQuery = { ...draft.queries[idx + 1] };
                    draft.queries[idx] = nextQuery;
                    draft.queries[idx + 1] = thisQuery;
                }
            };
        },

        moveQueryUp({id}) {
            return (draft:DataModelerState) => {
                const idx = draft.queries.findIndex((q) => q.id === id);
                if (idx > 0) {
                    const thisQuery = { ...draft.queries[idx] };
                    const nextQuery = { ...draft.queries[idx - 1] };
                    draft.queries[idx] = nextQuery;
                    draft.queries[idx - 1] = thisQuery;
                }
            }
        },
     }
 }