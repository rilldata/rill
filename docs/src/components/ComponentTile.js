import React from 'react';
import PropTypes from 'prop-types';
import { useHistory } from '@docusaurus/router';

/**
 * ComponentTile component for documentation tiles.
 * Displays a header, image, and link.
 * The entire tile is clickable and links to the specified URL.
 */
function ComponentTile({ header, image, link, multiple_measures = false }) {
    const history = useHistory();

    const handleClick = (e) => {
        // Check if the link contains an anchor
        if (link.includes('#')) {
            e.preventDefault();

            const [path, anchor] = link.split('#');

            // Use Docusaurus router to navigate
            history.push(link);

            // Try to scroll to the element after navigation
            setTimeout(() => {
                const element = document.getElementById(anchor);
                if (element) {
                    element.scrollIntoView({ behavior: 'smooth' });
                } else {
                    // Try alternative anchor formats
                    const alternatives = [
                        anchor,
                        anchor.replace(/-/g, ''),
                        header.toLowerCase(),
                        header.toLowerCase().replace(/\s+/g, '-')
                    ];

                    for (const alt of alternatives) {
                        const el = document.getElementById(alt);
                        if (el) {
                            el.scrollIntoView({ behavior: 'smooth' });
                            break;
                        }
                    }
                }
            }, 300);
        }
        // If no anchor, let the default behavior handle it
    };

    return (
        <a className="component-icon" href={link} onClick={handleClick}>
            <div className="component-icon-header">
                <span className="component-icon-header-text">{header}</span>
            </div>
            <div className="component-icon-content">
                {image && <div className="component-image">{image}</div>}
            </div>
            {multiple_measures && multiple_measures !== "False" && <div className="component-icon-multiple-measures">Supports Multiple Measures</div>}
        </a>
    );
}

ComponentTile.propTypes = {
    header: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    image: PropTypes.node,
    multiple_measures: PropTypes.bool,
};

export default ComponentTile; 