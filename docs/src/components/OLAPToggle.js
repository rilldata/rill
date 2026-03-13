import React, { useState, useEffect } from 'react';
import BrowserOnly from '@docusaurus/BrowserOnly';

/**
 * Toggle between DuckDB and ClickHouse content sections.
 * Uses the connector SVG logos as toggle buttons.
 * Syncs with the URL hash: #duckdb / #clickhouse.
 *
 * Usage in MDX:
 *   <OLAPToggle>
 *     <OLAPToggle.DuckDB>
 *       ...DuckDB content...
 *     </OLAPToggle.DuckDB>
 *     <OLAPToggle.ClickHouse>
 *       ...ClickHouse content...
 *     </OLAPToggle.ClickHouse>
 *   </OLAPToggle>
 */

const ENGINES = ['duckdb', 'clickhouse'];

function engineFromHash(hash) {
  const value = (hash || '').replace('#', '').toLowerCase();
  return ENGINES.includes(value) ? value : null;
}

function OLAPToggleInner({ children, defaultEngine = 'duckdb' }) {
  const [active, setActive] = useState(() => {
    return engineFromHash(window.location.hash) || defaultEngine;
  });

  // Listen for hash changes (back/forward, manual edits)
  useEffect(() => {
    function onHashChange() {
      const engine = engineFromHash(window.location.hash);
      if (engine) setActive(engine);
    }
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, []);

  function selectEngine(engine) {
    setActive(engine);
    // Update hash without triggering a scroll jump
    history.replaceState(null, '', `#${engine}`);
  }

  // Extract DuckDB and ClickHouse children
  let duckdbContent = null;
  let clickhouseContent = null;

  React.Children.forEach(children, (child) => {
    if (!child || !child.type) return;
    if (child.type === OLAPToggle.DuckDB || child.type.displayName === 'DuckDB') {
      duckdbContent = child.props.children;
    } else if (child.type === OLAPToggle.ClickHouse || child.type.displayName === 'ClickHouse') {
      clickhouseContent = child.props.children;
    }
  });

  return (
    <div className="olap-toggle" id="olap-toggle">
      <div className="olap-toggle__tabs" role="tablist">
        <button
          role="tab"
          aria-selected={active === 'duckdb'}
          className={`olap-toggle__tab ${active === 'duckdb' ? 'olap-toggle__tab--active' : ''}`}
          onClick={() => selectEngine('duckdb')}
        >
          <img
            src="/img/build/connectors/icons/Logo-DuckDB.svg"
            alt="DuckDB"
            className=""
          />
        </button>
        <button
          role="tab"
          aria-selected={active === 'clickhouse'}
          className={`olap-toggle__tab ${active === 'clickhouse' ? 'olap-toggle__tab--active' : ''}`}
          onClick={() => selectEngine('clickhouse')}
        >
          <img
            src="/img/build/connectors/icons/Logo-ClickHouse.svg"
            alt="ClickHouse"
            className="olap-toggle__logo"
          />
        </button>
      </div>
      <div className="olap-toggle__content" role="tabpanel">
        {active === 'duckdb' ? duckdbContent : clickhouseContent}
      </div>
    </div>
  );
}

function OLAPToggle(props) {
  return (
    <BrowserOnly fallback={<div className="olap-toggle">{props.children}</div>}>
      {() => <OLAPToggleInner {...props} />}
    </BrowserOnly>
  );
}

// Sub-components for content slots
OLAPToggle.DuckDB = function DuckDB({ children }) {
  return <>{children}</>;
};
OLAPToggle.DuckDB.displayName = 'DuckDB';

OLAPToggle.ClickHouse = function ClickHouse({ children }) {
  return <>{children}</>;
};
OLAPToggle.ClickHouse.displayName = 'ClickHouse';

export default OLAPToggle;
