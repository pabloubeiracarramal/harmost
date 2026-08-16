import { createFileRoute } from '@tanstack/react-router';
import { TokensPage } from '@/pages/tokens';

export const Route = createFileRoute('/_authenticated/tokens')({
  component: TokensPage,
});
