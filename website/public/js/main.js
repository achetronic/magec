import { Centella } from './centella.js';

document.addEventListener('DOMContentLoaded', () => {
  // Nav scroll
  const nav = document.querySelector('.nav');
  if (nav) window.addEventListener('scroll', () => nav.classList.toggle('scrolled', scrollY > 20), { passive: true });

  // Mobile nav toggle
  const toggle = document.querySelector('.nav__toggle');
  const links = document.querySelector('.nav__links');
  if (toggle && links) {
    toggle.addEventListener('click', () => links.classList.toggle('open'));
    links.querySelectorAll('a').forEach(l => l.addEventListener('click', () => links.classList.remove('open')));
  }

  // Hero orb
  const canvas = document.getElementById('hero-orb');
  if (canvas) new Centella(canvas).start();

  // Scroll reveal
  const obs = new IntersectionObserver(entries => {
    entries.forEach(e => { if (e.isIntersecting) e.target.classList.add('visible'); });
  }, { threshold: .1, rootMargin: '0px 0px -40px 0px' });
  document.querySelectorAll('.reveal, .stagger-children').forEach(el => obs.observe(el));

  // Docs sidebar active tracking
  const sidebar = document.querySelector('.docs-sidebar');
  if (sidebar) {
    const sidebarLinks = sidebar.querySelectorAll('.docs-sidebar__link');
    const sections = [];
    sidebarLinks.forEach(link => {
      const href = link.getAttribute('href');
      if (href?.startsWith('#')) {
        const sec = document.querySelector(href);
        if (sec) sections.push({ link, section: sec });
      }
    });
    if (sections.length) {
      const sObs = new IntersectionObserver(entries => {
        entries.forEach(e => {
          if (e.isIntersecting) {
            sidebarLinks.forEach(l => l.classList.remove('docs-sidebar__link--active'));
            const m = sections.find(s => s.section === e.target);
            if (m) m.link.classList.add('docs-sidebar__link--active');
          }
        });
      }, { threshold: .2, rootMargin: '-80px 0px -60% 0px' });
      sections.forEach(s => sObs.observe(s.section));
    }
    const sToggle = document.querySelector('.docs-sidebar-toggle');
    if (sToggle) sToggle.addEventListener('click', () => sidebar.classList.toggle('open'));
  }

  // Lightbox
  const lightbox = document.getElementById('lightbox');
  const lightboxImg = document.getElementById('lightbox-img');
  if (lightbox && lightboxImg) {
    document.querySelectorAll('.screenshot').forEach(img => {
      img.addEventListener('click', () => {
        lightboxImg.src = img.src;
        lightboxImg.alt = img.alt;
        lightbox.classList.add('active');
      });
    });
    lightbox.addEventListener('click', () => lightbox.classList.remove('active'));
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape') lightbox.classList.remove('active');
    });
  }
});
