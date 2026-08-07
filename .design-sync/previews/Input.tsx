import { Input } from 'front';

// Previews are static renders, so these are controlled inputs with a no-op
// onChange — the value shown is the value passed.
const noop = () => undefined;

const stack: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 14,
  maxWidth: 460,
};

/** The dispatch form's controls: mono for image refs, plain for prose. */
export function DispatchForm() {
  return (
    <div style={stack}>
      <Input value="ghcr.io/acme/api-tests:1.4.2" onChange={noop} mono required />
      <Input value="pytest -q --maxfail=1" onChange={noop} mono />
      <Input value="Nightly regression run" onChange={noop} />
    </div>
  );
}

/** Empty with placeholder — the resting state of the form. */
export function Placeholder() {
  return (
    <div style={stack}>
      <Input value="" onChange={noop} placeholder="ghcr.io/acme/api-tests:1.4.2" mono />
      <Input value="" onChange={noop} placeholder="Optional description" />
    </div>
  );
}

/** `mono` off vs on — same control, different face. */
export function MonoVsProse() {
  return (
    <div style={stack}>
      <Input value="Nightly regression run" onChange={noop} />
      <Input value="ghcr.io/acme/api-tests:1.4.2" onChange={noop} mono />
    </div>
  );
}

/** `type` passes straight through to the underlying input. */
export function Types() {
  return (
    <div style={stack}>
      <Input value="600" onChange={noop} type="number" />
      <Input value="hmt_a1b2c3d4e5f6" onChange={noop} type="password" mono />
    </div>
  );
}
