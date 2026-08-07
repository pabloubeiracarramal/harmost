import { Field, Input } from 'front';

const noop = () => undefined;

const stack: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 20,
  maxWidth: 460,
};

/** Field wraps the control in its <label>, so no htmlFor/id wiring is needed. */
export function WithInput() {
  return (
    <div style={stack}>
      <Field label="Image" hint="Fully-qualified image reference" required>
        <Input value="ghcr.io/acme/api-tests:1.4.2" onChange={noop} mono />
      </Field>
    </div>
  );
}

/** The dispatch form as a whole: required and optional fields together. */
export function DispatchForm() {
  return (
    <div style={stack}>
      <Field label="Image" hint="Fully-qualified image reference" required>
        <Input value="ghcr.io/acme/api-tests:1.4.2" onChange={noop} mono />
      </Field>
      <Field label="Command" hint="Overrides the image ENTRYPOINT">
        <Input value="pytest -q --maxfail=1" onChange={noop} mono />
      </Field>
      <Field label="Timeout" hint="Seconds before the job is killed">
        <Input value="600" onChange={noop} type="number" />
      </Field>
    </div>
  );
}

/** `required` adds the red asterisk; without it the label is plain. */
export function RequiredVsOptional() {
  return (
    <div style={stack}>
      <Field label="Image" required>
        <Input value="" onChange={noop} placeholder="ghcr.io/acme/api-tests:1.4.2" mono />
      </Field>
      <Field label="Description">
        <Input value="" onChange={noop} placeholder="Optional" />
      </Field>
    </div>
  );
}

/** Hint text is optional — omit it and the control sits tight under the label. */
export function NoHint() {
  return (
    <div style={stack}>
      <Field label="Agent">
        <Input value="build-runner-01" onChange={noop} />
      </Field>
    </div>
  );
}
