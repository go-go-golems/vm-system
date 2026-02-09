import type { EndpointBuilder } from '@reduxjs/toolkit/query';
import type { VmBaseQuery } from '../transport';

export type VmTagType = 'Template' | 'Session' | 'Execution';

export const VM_TAG_TYPES: VmTagType[] = ['Template', 'Session', 'Execution'];

export type VmEndpointBuilder = EndpointBuilder<VmBaseQuery, VmTagType, 'vmApi'>;

interface UiSliceState {
  sessionNames: Record<string, string>;
  currentSessionId: string | null;
}

export interface ApiRootState {
  ui: UiSliceState;
}
