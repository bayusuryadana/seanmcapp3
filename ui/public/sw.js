const CACHE_NAME = 'sean-wallet-v1'

self.addEventListener('install', (event) => {
  event.waitUntil(
    fetch('/')
      .then((response) => response.text())
      .then((html) => {
        const shellUrls = [...html.matchAll(/(?:src|href)="([^"]+)"/g)]
          .map((match) => new URL(match[1], self.location.origin))
          .filter((url) => url.origin === self.location.origin)
          .map((url) => url.pathname)
        return caches.open(CACHE_NAME).then((cache) => cache.addAll(['/', ...shellUrls]))
      }),
  )
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))),
    ).then(() => self.clients.claim()),
  )
})

self.addEventListener('fetch', (event) => {
  const request = event.request
  const url = new URL(request.url)

  // Never store private API responses on the device cache.
  if (url.origin !== self.location.origin || url.pathname.startsWith('/api/')) return

  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const copy = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(request, copy))
          return response
        })
        .catch(() => caches.match(request).then((cached) => cached || caches.match('/'))),
    )
    return
  }

  event.respondWith(
    caches.match(request).then((cached) =>
      cached || fetch(request).then((response) => {
        if (response.ok) {
          const copy = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(request, copy))
        }
        return response
      }),
    ),
  )
})
