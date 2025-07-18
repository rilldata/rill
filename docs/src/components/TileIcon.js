import React from 'react';
import PropTypes from 'prop-types';

/**
 * TileIcon component for documentation tiles.
 * Displays a header and content.
 * Main card links to the specified destination.
 */
function TileIcon({ header, content, link, linkLabel = '', target, rel }) {
    return (
        <a className="tile-icon" href={link} target={target} rel={rel}>
            <div className="tile-icon-header">{header}</div>
            <div className="tile-icon-content">
                {content}
            </div>
            <div className="tile-icon-footer">
                <span className={`tile-icon-link-right ${linkLabel ? 'with-arrow' : 'no-arrow'}`}>
                    {linkLabel}
                </span>
            </div>
        </a>
    );
}

TileIcon.propTypes = {
    header: PropTypes.string.isRequired,
    content: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    linkLabel: PropTypes.string,
    target: PropTypes.string,
    rel: PropTypes.string,
};

export default TileIcon; 