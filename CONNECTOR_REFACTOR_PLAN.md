# 🔄 Connector Explorer Refactor Plan

## **Current State vs Target State**

### **Before (OLAP-only):**

```
ConnectorExplorer
├── {#if implementsOlap}
│   └── ConnectorEntry (OLAP only)
│       └── DatabaseExplorer (uses OLAPListTables)
│           ├── DatabaseEntry
│           ├── DatabaseSchemaEntry
│           └── TableEntry
```

### **After (All Connector Types):**

```
ConnectorExplorer
├── ConnectorEntry (All connectors)
│   └── DatabaseExplorer (auto-detects API)
│       ├── DatabaseEntry (hybrid API support)
│       ├── DatabaseSchemaEntry (hybrid API support)
│       └── TableEntry (hybrid API support)
```

---

## **✅ Phase 1: Foundation (COMPLETED)**

### **New Selectors Created** (`web-common/src/features/connectors/selectors.ts`)

- ✅ `useDatabaseSchemas()` - Uses `ListDatabaseSchemas` API
- ✅ `useDatabasesFromSchemas()` - Extracts unique databases
- ✅ `useSchemasForDatabase()` - Filters schemas by database
- ✅ `useTablesForSchema()` - Uses `ListTables` API (on-demand)
- ✅ `useTableMetadata()` - Uses `GetTable` API (on-demand)
- ✅ `useConnectorCapabilities()` - Detects connector type
- ✅ Compatibility layer for existing API shapes

### **Hybrid Components Created**

- ✅ `DatabaseExplorer.svelte` - Auto-detects API based on connector type
- ✅ `DatabaseEntry.svelte` - Supports both API approaches

---

## **🚧 Phase 2: Complete Component Migration**

### **Components to Update:**

#### **1. DatabaseSchemaEntry.svelte**

```typescript
// NEEDED: Update to use hybrid API approach
export let useNewAPI: boolean = false;

$: tablesQuery = useNewAPI
  ? useTablesForSchema(instanceId, connectorName, database, databaseSchema)
  : useTablesLegacy(instanceId, connectorName, database, databaseSchema);
```

#### **2. TableEntry.svelte**

```typescript
// NEEDED: Handle V1TableInfo vs V1OlapTableInfo differences
export let tableInfo: V1TableInfo | V1OlapTableInfo;
export let useNewAPI: boolean = false;

// New API doesn't have hasUnsupportedDataTypes, physicalSizeBytes
$: hasUnsupportedDataTypes = useNewAPI
  ? false
  : tableInfo.hasUnsupportedDataTypes;
```

#### **3. ConnectorEntry.svelte**

```typescript
// NEEDED: Remove OLAP-only restriction
// Change from:
{#if implementsOlap}

// To:
{#if implementsOlap || implementsSqlStore}
  <DatabaseExplorer {instanceId} {connector} {store} />
{/if}
```

---

## **🎯 Phase 3: API Consolidation Strategy**

### **Connector Type Detection Logic:**

```typescript
// Decision tree for API selection
function getApiStrategy(connector: V1AnalyzedConnector) {
  const { implementsOlap, implementsSqlStore } = connector.driver;

  if (implementsOlap) {
    return "legacy"; // Continue using OLAPListTables for better performance
  } else if (implementsSqlStore) {
    return "new"; // Use ListDatabaseSchemas → ListTables → GetTable
  } else {
    return "none"; // Don't show in explorer
  }
}
```

### **Performance Considerations:**

- **OLAP Connectors**: Keep using `OLAPListTables` (single call, better performance)
- **SQL Connectors**: Use new granular APIs (necessary for non-OLAP)
- **Lazy Loading**: Only call `ListTables` when schema is expanded

---

## **📊 Data Structure Mapping**

### **API Response Differences:**

| Field                     | OLAPListTables | ListTables | Notes                     |
| ------------------------- | -------------- | ---------- | ------------------------- |
| `database`                | ✅             | ❌         | Passed as parameter       |
| `databaseSchema`          | ✅             | ❌         | Passed as parameter       |
| `name`                    | ✅             | ✅         | Table name                |
| `hasUnsupportedDataTypes` | ✅             | ❌         | Need GetTable for details |
| `physicalSizeBytes`       | ✅             | ❌         | Need GetTable for details |
| `view`                    | ❌             | ✅         | New field                 |

### **Component Props Updates Needed:**

```typescript
// TableEntry.svelte - Handle missing fields gracefully
export let hasUnsupportedDataTypes: boolean = false; // Default for new API
export let physicalSizeBytes: number = -1; // Default for new API
export let view: boolean = false; // New field from new API
```

---

## **🧪 Phase 4: Testing Strategy**

### **Test Cases:**

1. **OLAP Connectors** (DuckDB, ClickHouse) - Should use legacy API
2. **SQL Connectors** (Postgres, MySQL) - Should use new API
3. **Mixed Projects** - Both connector types should work
4. **Error Handling** - Network failures, empty results
5. **Loading States** - Each hierarchy level loads independently

### **Feature Flags:**

```typescript
// Optional: Add feature flag for gradual rollout
const { useNewConnectorAPIs } = featureFlags;
$: shouldUseNewAPI =
  useNewConnectorAPIs || (!implementsOlap && implementsSqlStore);
```

---

## **⚡ Phase 5: Performance Optimizations**

### **Caching Strategy:**

- **Database Schemas**: Cache at connector level (changes rarely)
- **Tables**: Cache at schema level, invalidate on refresh
- **Table Metadata**: Cache per table, lazy load

### **Loading UX:**

```typescript
// Stagger loading states for better UX
{#if isLoading}
  <div class="loading-skeleton">
    <div class="skeleton-line"></div>
    <div class="skeleton-line short"></div>
  </div>
{/if}
```

---

## **🚀 Phase 6: Final Migration**

### **Rollout Steps:**

1. **Feature Flag**: Enable for internal testing
2. **Gradual Rollout**: Enable for SQL connectors first
3. **Monitor**: Check error rates, performance
4. **Full Migration**: Remove legacy code paths
5. **Cleanup**: Remove old `/olap/` selectors

### **Breaking Changes:**

- ⚠️ `V1OlapTableInfo` → `V1TableInfo` in some contexts
- ⚠️ Missing fields (`hasUnsupportedDataTypes`, `physicalSizeBytes`) for SQL connectors
- ⚠️ Component prop changes for `useNewAPI` parameter

---

## **🎉 Benefits After Migration**

1. **🔗 Universal Connector Support** - Postgres, MySQL, etc. now browsable
2. **⚡ Efficient Loading** - On-demand API calls instead of bulk fetch
3. **🏗️ Better Architecture** - Separation of concerns between connector types
4. **🚀 Future Ready** - Easy to add new connector types
5. **🐛 Better Error Handling** - Granular error states per hierarchy level

---

## **📋 Implementation Checklist**

### **Phase 2 - Component Updates:**

- [ ] Update `DatabaseSchemaEntry.svelte` with hybrid API support
- [ ] Update `TableEntry.svelte` to handle both data structures
- [ ] Update `ConnectorEntry.svelte` to show all connector types
- [ ] Create `TableSchema.svelte` variant for new API

### **Phase 3 - Testing:**

- [ ] Add unit tests for new selectors
- [ ] Add integration tests for mixed connector scenarios
- [ ] Test error handling and loading states

### **Phase 4 - Rollout:**

- [ ] Add feature flag
- [ ] Deploy with flag disabled
- [ ] Enable for SQL connectors
- [ ] Monitor metrics
- [ ] Full rollout

### **Phase 5 - Cleanup:**

- [ ] Remove old OLAP-only restrictions
- [ ] Update documentation
- [ ] Remove unused legacy code
- [ ] Update TypeScript types

---

## **🔧 Technical Debt & Future Improvements**

1. **Type Safety**: Create discriminated unions for `TableInfo` variants
2. **Error Boundaries**: Add React-style error boundaries for each component
3. **Virtualization**: For connectors with thousands of tables
4. **Search**: Add table search across all connector types
5. **Bookmarks**: Save frequently accessed tables
