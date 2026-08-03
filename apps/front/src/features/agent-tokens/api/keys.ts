export const agentTokenKeys = {
  all: ['agent-tokens'] as const,
  lists: () => [...agentTokenKeys.all, 'list'] as const,
  list: () => [...agentTokenKeys.lists()] as const,
};
