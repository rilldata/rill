import { useEffect } from 'react';
import { useLocation } from '@docusaurus/router';

/**
 * HubSpot SPA pageview tracking for Docusaurus
 * Tracks route changes and sends pageview events to HubSpot
 */
function HubSpotPageViews() {
  const location = useLocation();

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const _hsq = (window._hsq = window._hsq || []);
    const path = location.pathname + location.search + location.hash;

    _hsq.push(['setPath', path]);
    _hsq.push(['trackPageView']);
  }, [location]);

  return null;
}

export default HubSpotPageViews;
