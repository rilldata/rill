import React from 'react';
import PropTypes from 'prop-types';

/**
 * ComponentTile component for documentation tiles.
 * Displays a header, image, and link.
 * The entire tile is clickable and links to the specified URL.
 */
function ComponentTile({ header, image, link }) {

    return (
        <a className="component-icon" href={link}>
            <div className="component-icon-header">
                <span className="component-icon-header-text">{header}</span>
            </div>
            <div className="component-icon-content">
                {image && <div className="component-icon-header-icon-top">{image}</div>}
            </div>
        </a>
    );
}

ComponentTile.propTypes = {
    header: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    image: PropTypes.node,
};

export default ComponentTile; 