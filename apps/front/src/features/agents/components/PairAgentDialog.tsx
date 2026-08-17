import { useEffect, useState } from 'react';
import { useApproveDevice } from '../api/mutations';
import { Button } from '@/shared/components/ui/button';
import { Input } from '@/shared/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/components/ui/dialog';

interface PairAgentDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Pre-fills the code, e.g. when arriving via the `?code=` link `agent pair` prints. */
  initialCode?: string;
}

export function PairAgentDialog({
  open,
  onOpenChange,
  initialCode,
}: PairAgentDialogProps) {
  const [code, setCode] = useState(initialCode ?? '');
  const approve = useApproveDevice();

  // Reset on every open so a previous attempt's error/code doesn't linger.
  useEffect(() => {
    if (open) {
      setCode(initialCode ?? '');
      approve.reset();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, initialCode]);

  const submit = () => {
    const trimmed = code.trim();
    if (!trimmed) return;
    approve.mutate(trimmed, { onSuccess: () => onOpenChange(false) });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Pair new agent</DialogTitle>
          <DialogDescription>
            Run{' '}
            <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs">
              harmost pair &lt;hub-url&gt;
            </code>{' '}
            on the machine you want to connect, then enter the code it prints.
          </DialogDescription>
        </DialogHeader>

        <Input
          autoFocus
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder="XXXX-XXXX"
          maxLength={9}
          className="text-center font-mono text-lg tracking-widest uppercase"
        />
        {approve.isError && (
          <p className="text-sm text-red-400">
            {(approve.error as Error).message}
          </p>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={approve.isPending || !code.trim()}>
            {approve.isPending ? 'Pairing…' : 'Pair agent'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
