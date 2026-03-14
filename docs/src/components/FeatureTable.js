import React from 'react';
import PropTypes from 'prop-types';

function FeatureTable({ columns, rows }) {
  return (
    <div className="feature-table-wrapper">
      <div className="feature-table">
        <div className="feature-table-header">
          {columns.map((col, i) => (
            <div
              key={i}
              className={`feature-table-cell ${i === 0 ? 'feature-table-label-col' : 'feature-table-value-col'}`}
            >
              {col}
            </div>
          ))}
        </div>
        {rows.map((row, i) => (
          <div key={i} className="feature-table-row">
            {row.map((cell, j) => (
              <div
                key={j}
                className={`feature-table-cell ${j === 0 ? 'feature-table-label-col' : 'feature-table-value-col'}`}
              >
                {cell === true ? (
                  <span className="feature-table-check">&#10003;</span>
                ) : cell === false ? (
                  <span className="feature-table-dash">&mdash;</span>
                ) : (
                  cell
                )}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

FeatureTable.propTypes = {
  columns: PropTypes.arrayOf(PropTypes.string).isRequired,
  rows: PropTypes.arrayOf(PropTypes.array).isRequired,
};

export default FeatureTable;
