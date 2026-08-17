import { createFileRoute } from '@tanstack/react-router';
import { NewJobPage } from '@/pages/jobs/new';

export const Route = createFileRoute('/_authenticated/jobs/new')({
  component: NewJobPage,
});
