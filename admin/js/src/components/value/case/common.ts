import { ParamType } from '../../../api';

export const caseConditions = ['server', 'group', 'datacenter', 'service'] as const;
export type CaseConditions = typeof caseConditions[number];

// a branch is never encrypted: one that needs a secret is a symlink to an
// encrypted parameter
export interface Case extends Partial<Record<CaseConditions, string>> {
	mime: ParamType;
	value: string | null;
}
