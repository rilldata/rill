/**
 * The type definition for a "profile column"
 */
 export interface ProfileColumn {
    name: string;
    type: string;
    conceptualType: string;
    summary?: (CategoricalSummary | any); // FIXME
    nullCount?:number;
}

export interface Item {
    id: string;
}

export interface Query extends Item {
    /**  */
    query: string;
    /** sanitizedQuery is always a 1:1 function of the query itself */
    sanitizedQuery: string;
    /** name is used for the filename and exported file */
    name: string;
    /** cardinality is the total number of rows of the previewed dataset */
    cardinality?: number;
    /** sizeInBytes is the total size of the previewed dataset. 
     * It is not generated until the user exports the query.
     * */
    sizeInBytes?: number; // TODO: make sure this is just size
    error?: string;
    sources?: string[];
    profile?: ProfileColumn[]; // TODO: create Profile interface
    preview?: any;
    destinationProfile?: any;
}

export interface Source extends Item {
    id: string;
    path: string;
    name: string;
    profile: ProfileColumn[]; // TODO: create Profile interface
    head: any[];
    cardinality?: number;
    sizeInBytes?: number;
    nullCounts?:any;
}

export interface CategoricalSummary {
    topK:TopKEntry[];
    cardinality:number;
}

export interface NumericSummary {
    histogram:NumericHistogramBin[]
}

export interface TopKEntry {
    value:any;
    count:number;
}

export interface NumericHistogramBin {
    bucket:number;
    low:number;
    high:number;
    count:number
}

/**
 * The entire state object for the data modeler.
 */
export interface DataModelerState {
    activeQuery?: string;
    queries: Query[];
    sources: Source[];
    status: string;
}