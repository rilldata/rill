// src/components/ComingSoon.js
import React, { useEffect } from 'react';

const ComingSoon = () => {
  useEffect(() => {
    const content = document.getElementById('contents_to_overlay');
    if (content) {
      content.innerHTML = '<div style="display: flex; justify-content: center; align-items: center; height: 150px; font-size: 2rem;">Coming Soon!</div>';
    }
  }, []);

  return null;
};

export default ComingSoon;
