import React from 'react';
import PropTypes from 'prop-types';

function FeatureList({ items }) {
  return (
    <div className="feature-list">
      {items.map((item, index) => (
        <a key={index} className="feature-list-item" href={item.link}>
          <span className="feature-list-name">{item.name}</span>
          <span className="feature-list-desc">{item.description}</span>
        </a>
      ))}
    </div>
  );
}

FeatureList.propTypes = {
  items: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      description: PropTypes.string.isRequired,
      link: PropTypes.string.isRequired,
    })
  ).isRequired,
};

export default FeatureList;
