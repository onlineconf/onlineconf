import { ParamType, ValueFlags } from '../../../api';

export const caseConditions = ['server', 'group', 'datacenter', 'service'] as const;
export type CaseConditions = typeof caseConditions[number];

// the encrypted flag is present (true) only on encrypted branches, so plain
// case values keep their historical JSON shape byte for byte
export interface Case extends Partial<Record<CaseConditions, string>>, ValueFlags {
	mime: ParamType;
	value: string | null;
}
