---
category: Jobs
---
Field — label, optional hint and required marker wrapped around a form control.

```tsx
<Field label="Image" hint="Fully-qualified image reference" required>
  <Input value={image} onChange={setImage} placeholder="ghcr.io/acme/api-tests:1.4.2" mono />
</Field>
```

Renders a `<label>` element wrapping its `children`, so the control inside is associated with the label without any `htmlFor`/`id` wiring. `required` appends a red asterisk to the label; `hint` renders dim helper text underneath the control.

It owns only the labelling — sizing and validation styling belong to the control you pass as children.
