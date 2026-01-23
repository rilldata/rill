import { useColorMode } from '@docusaurus/theme-common';
import { useLocation } from '@docusaurus/router';
import Navbar from '@theme-original/Navbar';
import { useEffect, useLayoutEffect } from 'react';

// Mobile icon functionality removed

export default function NavbarWrapper(props) {
  // We only need colorMode to determine which icon to show.
  // The toggle logic is handled by the original button's onClick.
  const { colorMode } = useColorMode();
  const location = useLocation();

  // Close all open dropdowns when route changes
  // This fixes the issue where hover-based dropdowns stay open after clicking an item
  useEffect(() => {
    const closeAllDropdowns = () => {
      // Remove dropdown--show class from all dropdowns
      const openDropdowns = document.querySelectorAll('.navbar__item.dropdown.dropdown--show');
      openDropdowns.forEach((dropdown) => {
        dropdown.classList.remove('dropdown--show');
      });

      // Also reset aria-expanded attributes
      const expandedLinks = document.querySelectorAll('.navbar__link[aria-expanded="true"]');
      expandedLinks.forEach((link) => {
        link.setAttribute('aria-expanded', 'false');
      });

      // Blur any focused navbar elements to prevent hover state from re-triggering
      const activeElement = document.activeElement;
      if (activeElement && activeElement.closest('.navbar__item.dropdown')) {
        activeElement.blur();
      }
    };

    closeAllDropdowns();
  }, [location.pathname]);

  // Handle Dark Mode Toggle Icons
  // useLayoutEffect fires synchronously before paint, reducing flicker
  useLayoutEffect(() => {
    const toggleButtons = document.querySelectorAll('button[class*="toggleButton"]');
    if (!toggleButtons.length) return;

    toggleButtons.forEach((btn, index) => {
      // Add ID to first toggle for potential targeting
      if (index === 0) btn.id = 'dark-mode-toggle';

      // Find or create the container for our custom icon
      let iconContainer = btn.querySelector('.icon-container');
      if (!iconContainer) {
        iconContainer = document.createElement('span');
        iconContainer.className = 'icon-container';

        // Clear existing Docusaurus toggle content (text/emojis)
        btn.innerHTML = '';
        btn.appendChild(iconContainer);
      }

      // Update Icon based on CURRENT mode (show the OPPOSITE icon)
      // If Dark Mode -> Show Sun (to switch to Light)
      // If Light Mode -> Show Moon (to switch to Dark)
      const isDark = colorMode === 'dark';

      iconContainer.innerHTML = `
        <img 
          src="/icons/${isDark ? 'Sun' : 'Moon'}.svg" 
          alt="${isDark ? 'Switch to light mode' : 'Switch to dark mode'}" 
          width="24" 
          height="24" 
        />
      `;

      btn.setAttribute('aria-label', isDark ? 'Switch to light mode' : 'Switch to dark mode');
    });
  }, [colorMode]);

  // Mobile sidebar icons removed - no longer needed

  // Consolidated MutationObserver for navbar updates
  useEffect(() => {
    const markActiveDropdowns = () => {
      // Target elements with both navbar__item and dropdown classes
      const dropdownItems = document.querySelectorAll('.navbar__item.dropdown');
      dropdownItems.forEach((dropdownItem) => {
        const activeDropdownLink = dropdownItem.querySelector('.dropdown__link--active');
        if (activeDropdownLink) {
          dropdownItem.classList.add('navbar__dropdown--has-active');
        } else {
          dropdownItem.classList.remove('navbar__dropdown--has-active');
        }
      });
    };

    const markActiveNavItems = () => {
      // Target non-dropdown navbar items
      const navItems = document.querySelectorAll('.navbar__item:not(.dropdown)');
      navItems.forEach((navItem) => {
        const activeLink = navItem.querySelector('.navbar__link--active');
        if (activeLink) {
          navItem.classList.add('navbar__item--has-active-link');
        } else {
          navItem.classList.remove('navbar__item--has-active-link');
        }
      });
    };

    const addDataTextAttributes = () => {
      const navLinks = document.querySelectorAll('.navbar__link');
      navLinks.forEach((link) => {
        // Only add if not already present
        if (!link.hasAttribute('data-text')) {
          const text = link.textContent?.trim() || '';
          if (text) {
            link.setAttribute('data-text', text);
          }
        }
      });
    };

    const markCustomDropdowns = () => {
      // Add class hook for dropdowns with custom dropdown links (replaces :has(.my-custom-dropdown))
      const customDropdownLinks = document.querySelectorAll('.navbar__link.my-custom-dropdown');
      customDropdownLinks.forEach((link) => {
        const dropdownItem = link.closest('.navbar__item.dropdown');
        if (dropdownItem) {
          dropdownItem.classList.add('dropdown--custom');
        }
      });
    };

    const markIconLinkItems = () => {
      // Add class hook for navbar items containing icon links (replaces :has(.navbar-icon-link))
      const iconLinks = document.querySelectorAll('.navbar-icon-link');
      iconLinks.forEach((link) => {
        const navItem = link.closest('.navbar__item');
        if (navItem) {
          navItem.classList.add('navbar__item--has-icon-link');
        }
        // Also mark parent list items in mobile sidebar
        const menuListItem = link.closest('.menu__list-item');
        if (menuListItem) {
          menuListItem.classList.add('menu__list-item--has-icon-link');
        }
        // Mark sidebar items
        const sidebarItem = link.closest('.navbar-sidebar__item');
        if (sidebarItem) {
          sidebarItem.classList.add('navbar-sidebar__item--has-icon-link');
        }
      });
    };

    const replaceCustomDropdownCarets = () => {
      // Add custom SVG chevron for custom dropdown links
      const customDropdownLinks = document.querySelectorAll('.navbar__link.my-custom-dropdown');
      customDropdownLinks.forEach((link) => {
        const dropdownItem = link.closest('.navbar__item.dropdown');
        if (dropdownItem && !link.hasAttribute('data-custom-chevron-added')) {
          // Mark as processed
          link.setAttribute('data-custom-chevron-added', 'true');

          // Create a container for the custom chevron
          let chevronContainer = link.querySelector('.custom-chevron');
          if (!chevronContainer) {
            chevronContainer = document.createElement('span');
            chevronContainer.className = 'custom-chevron';
            link.appendChild(chevronContainer);
          }

          // Clear and add custom SVG chevron
          chevronContainer.innerHTML = '';
          const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
          svg.setAttribute('height', '14px');
          svg.setAttribute('viewBox', '0 0 24 24');
          svg.setAttribute('fill', 'currentColor');
          svg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');

          const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
          path.setAttribute('fill-rule', 'evenodd');
          path.setAttribute('clip-rule', 'evenodd');
          path.setAttribute('d', 'M19.189 9.43683C19.3842 9.63209 19.3842 9.94867 19.189 10.1439L11.9999 17.333L4.81075 10.1439C4.61549 9.94867 4.61549 9.63209 4.81075 9.43683L5.98898 8.2586C6.18424 8.06334 6.50082 8.06334 6.69609 8.2586L11.9999 13.5624L17.3036 8.2586C17.4989 8.06334 17.8155 8.06334 18.0108 8.2586L19.189 9.43683Z');

          svg.appendChild(path);
          chevronContainer.appendChild(svg);
        }
      });
    };

    // Combined update function
    const updateNavbar = () => {
      markActiveDropdowns();
      markActiveNavItems();
      markCustomDropdowns();
      markIconLinkItems();
      addDataTextAttributes();
      replaceCustomDropdownCarets();
    };

    // Run on mount and when DOM changes
    updateNavbar();

    // Staggered timeouts to handle dynamically rendered navbar elements.
    // Docusaurus may hydrate or lazy-load navbar items at different times,
    // especially for dropdowns and client-side navigation. These delays
    // ensure our custom styling (data-text attrs, active states, chevrons)
    // is applied even if elements render after initial mount.
    setTimeout(updateNavbar, 100);
    setTimeout(updateNavbar, 500);
    setTimeout(updateNavbar, 1000);

    const observer = new MutationObserver(updateNavbar);
    observer.observe(document.body, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
    };
  }, []);

  return <Navbar {...props} />;
}
