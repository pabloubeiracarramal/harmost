---
category: Jobs
---
Input — the single-line text control used by the job dispatch form.

```tsx
<Input value={image} onChange={setImage} placeholder="ghcr.io/acme/api-tests:1.4.2" mono required />
```

`onChange` receives the **string value**, not the event — wire it straight to a `useState` setter. `mono` switches to a monospace face for image refs, commands and env values; `type` passes through to the underlying input (defaults to `text`).

Always full width (`w-full`), so control the size from the parent. Focus moves the border to indigo and removes the default outline. There is no built-in label — wrap it in `Field` for that.
