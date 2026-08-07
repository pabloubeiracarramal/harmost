export const jobKeys = {
  all: ['jobs'] as const,
  lists: () => [...jobKeys.all, 'list'] as const,
  list: () => [...jobKeys.lists()] as const,
  details: () => [...jobKeys.all, 'detail'] as const,
  detail: (id: string) => [...jobKeys.details(), id] as const,
  logs: (id: string) => [...jobKeys.detail(id), 'logs'] as const,
};
