# Frontend Conventions

## Rendering Model

Default:

```text
Browser
  ↓
HTTP request
  ↓
Go handler
  ↓
service
  ↓
repository
  ↓
templ
  ↓
HTML
  ↓
HTMX enhancement
  ↓
minimal vanilla JS
```

## Component Rules

Create a templ component when markup is:

- reused;
- stateful;
- semantically meaningful;
- complex enough to obscure a page.

Examples:

- `Button`
- `Input`
- `Modal`
- `Toast`
- `Pagination`
- `DataTable`
- `EmptyState`
- `FormField`

## HTMX Rules

Every HTMX interaction must define:

- request;
- target;
- swap;
- loading state;
- success state;
- error state.

Avoid deeply nested HTMX behavior that becomes difficult to reason about.

## JS Rules

Prefer event delegation and small modules.

Do not duplicate server-side business rules in JavaScript.

Client-side validation is for UX; server-side validation remains authoritative.

## Accessibility

Required:

- semantic HTML;
- labels for controls;
- keyboard navigation;
- visible focus;
- appropriate ARIA only where necessary;
- meaningful error messages;
- dialogs that manage focus;
- sufficient touch targets.

## Responsive Design

Design from small screens upward.

Check at minimum:

- mobile portrait;
- tablet;
- desktop;
- wide desktop.

Do not simply shrink desktop UI.

## PWA UX

Support:

- installable app shell;
- clear offline/online states;
- resilient loading;
- graceful service-worker updates;
- browser back/forward behavior.
