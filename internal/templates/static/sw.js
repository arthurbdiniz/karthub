const CACHE_NAME = 'karthub-v1';

self.addEventListener('install', (event) => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(clients.claim());
});

self.addEventListener('fetch', (event) => {
  // Network-first strategy for a dynamic app
  event.respondWith(
    fetch(event.request).catch(() => caches.match(event.request))
  );
});
