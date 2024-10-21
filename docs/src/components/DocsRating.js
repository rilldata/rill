import React, { useState } from 'react';
import { useLocation } from '@docusaurus/router';

const DocsRating = () => {
  const [voted, setVoted] = useState(false);
  const location = useLocation();

  const handleVote = (vote) => {
    if (!voted) {
      // Push the vote event to the GTM dataLayer
      if (window.dataLayer) {
        window.dataLayer.push({
          event: 'custom_vote',
          category: 'DocsRating',
          label: location.pathname,
          value: vote,
        });
      }
      setVoted(true);
    }
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: '10px',
      marginTop: '20px'
    }}>
      <p style={{ margin: 0 }}>Was this content helpful?</p>
      <button
        onClick={() => handleVote('up')}
        disabled={voted}
        aria-label="Thumbs up"
        style={{
          background: 'none',
          border: 'none',
          cursor: voted ? 'not-allowed' : 'pointer',
          fontSize: '1.5rem',
          opacity: voted ? 0.5 : 1,
          display: 'flex',
          alignItems: 'center',
        }}
      >
        👍
      </button>
      <button
        onClick={() => handleVote('down')}
        disabled={voted}
        aria-label="Thumbs down"
        style={{
          background: 'none',
          border: 'none',
          cursor: voted ? 'not-allowed' : 'pointer',
          fontSize: '1.5rem',
          opacity: voted ? 0.5 : 1,
          display: 'flex',
          alignItems: 'center',
        }}
      >
        👎
      </button>
    </div>
  );
};

export default DocsRating;
