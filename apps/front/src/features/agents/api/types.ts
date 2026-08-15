import type { components } from '@/shared/api/schema';

export type Agent = components['schemas']['Agent'];
export type ContainerInfo = components['schemas']['ContainerInfo'];
export type ContainerPort = components['schemas']['ContainerPort'];
export type ContainerMount = components['schemas']['ContainerMount'];
export type ContainerStats = components['schemas']['ContainerStats'];
export type ContainerActionPayload = components['schemas']['ContainerActionPayload'];
export type ContainerActionKind = ContainerActionPayload['action'];
