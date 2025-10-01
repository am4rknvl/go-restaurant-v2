# Restaurant System - Backend
## Deployment

### Local
```
docker-compose up -d --build
go run .
```

### Staging
```
docker build -t your-registry/restaurant-system:$(git rev-parse --short HEAD) .
docker push your-registry/restaurant-system:$(git rev-parse --short HEAD)
kubectl apply -f k8s/
```

### Production
```
kubectl set image deployment/restaurant-app app=your-registry/restaurant-system:$(git rev-parse --short HEAD)
kubectl rollout status deployment/restaurant-app
```
# Restaurant System - Backend

This repository contains a Go (Gin) backend for a restaurant table ordering platform. This patch added several endpoints and WebSocket support. Apply migrations in `migrations.sql` to your Postgres database.

Key routes added (examples):

- GET /api/v1/restaurant/:restaurant_id/table/:table_id/menu  -- customer menu grouped by category
- POST /api/v1/menu/item  -- create menu item (admin)
- PUT /api/v1/menu/item/:id
- DELETE /api/v1/menu/item/:id

- POST /api/v1/categories
- GET /api/v1/categories
- PUT /api/v1/categories/:id
- DELETE /api/v1/categories/:id

- POST /api/v1/tables
- GET /api/v1/tables
- GET /api/v1/tables/:id

- POST /api/v1/sessions
- GET /api/v1/sessions/:id
- PUT /api/v1/sessions/:id/close

- GET /api/v1/kitchen/orders
- PUT /api/v1/kitchen/orders/:id/status

- WebSocket endpoint: GET /ws?role=kitchen|customer|staff

Example JSON to create a menu item:

{
  "id": "uuid-if-you-want",
  "name": "Fried Rice",
  "description": "Delicious",
  "price": 5.5,
  "category": "Rice",
  "available": true,
  "image_url": "https://...",
  "special_notes": "No onions"
}

Notes:
- This change focuses on scaffolding endpoints and services; refine auth (JWT) and rate-limiting for production.
- Run `go build` to compile and fix any environment-specific issues.

Additional endpoints added in this update:

- Offline / Sync
  - POST /api/v1/orders/sync
    - Accepts { customer_id, session_id?, orders: [[{menu_item_id, quantity, special_instructions}, ...], ...] }
    - Used by offline-capable clients to batch-create orders when connectivity is restored.

- Notifications
  - POST /api/v1/notifications/subscribe  (body: {account_id, kind: 'push'|'sms', endpoint, metadata})
  - DELETE /api/v1/notifications/:id
  - GET /api/v1/notifications/account/:account_id
  - POST /api/v1/kitchen/notifications/send  (admin/staff trigger)

- Reservations
  - POST /api/v1/reservations  (create)
  - GET /api/v1/reservations  (list upcoming)
  - PUT /api/v1/reservations/:id  (update)
  - DELETE /api/v1/reservations/:id  (cancel)

- ETA / Prep
  - PUT /api/v1/orders/:id/eta  (body: {estimated_ready_at: RFC3339 timestamp})
    - Updates order ETA; broadcasts an "order_eta_updated" websocket event.

- Refunds & Partial Payments
  - POST /api/v1/payments/:id/refund  (body: {amount, reason})
  - POST /api/v1/payments/partial  (body: {order_id, amount})

Frontend helpers:
- web/static/sw.js  - Basic service worker that caches menu GET requests and supports a simple postMessage-based sync trigger.
- web/static/menu-sync.js - Minimal client utilities to register SW, cache menu in localStorage, queue orders and attempt sync on network restore.

Database migrations:
- Update your DB with `migrations.sql` to add tables: reservations, subscriptions, refunds and new columns for ETA and payment/refund support.

Security notes:
- Push & SMS stubs are placeholders; integrate a proper web-push library with VAPID keys and an SMS gateway (Twilio, Africa's Talking, etc.) for production.
- Ensure JWT secrets and provider credentials are stored in environment variables and not in source.
