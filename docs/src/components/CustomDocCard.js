import React from 'react';
import Link from '@docusaurus/Link';

// Define images or colors for specific documents


const CustomDocCard = ({ item }) => {
  // Get custom style (image or background color) based on docId

  return (
    <Link to={item.href} className="custom-doc-card">
      <div className="custom-doc-card__header">
        <h3>{item.label}</h3>
      </div>
    </Link>
  );
};

export default CustomDocCard;
