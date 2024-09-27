import React from 'react';
import Link from '@docusaurus/Link';

// Define images or colors for specific documents
const cardStyles = {
  'tutorials/rill_basics': {
    backgroundImage: 'url(/img/guide-image.jpg)',
    backgroundColor: '#f9f9f9',
  },
  'tutorials/rill_advanced_features/overview': {
    backgroundImage: 'url(/img/overview-image.jpg)',
    backgroundColor: '#eef5f9',
  },
  // Add more styles for different docs here
};

const CustomDocCard = ({ item }) => {
  // Get custom style (image or background color) based on docId
  const style = cardStyles[item.docId] || {};

  return (
    <Link to={item.href} className="custom-doc-card" style={style}>
      <div className="custom-doc-card__header">
        <h3>{item.label}</h3>
      </div>
    </Link>
  );
};

export default CustomDocCard;
