import * as React from 'react';
import { SvgIconProps } from '@mui/material/SvgIcon';

import DefaultIcon from '@mui/icons-material/Settings';
import RootIcon from '@mui/icons-material/Language';
import TopLevelIcon from '@mui/icons-material/Domain';
import NullIcon from '@mui/icons-material/Block';
import FolderIcon from '@mui/icons-material/FolderOpen';
import TextIcon from '@mui/icons-material/Subject';
import StructIcon from '@mui/icons-material/DeviceHub';
import SymlinkIcon from '@mui/icons-material/Launch';
import CaseIcon from '@mui/icons-material/Help';
import TemplateIcon from '@mui/icons-material/LocalAtm';
import ListIcon from '@mui/icons-material/List';

import { ParamType } from '../api';

const iconByType: { [P in ParamType]: React.ComponentType<SvgIconProps> } = {
	'application/x-null': NullIcon,
	'text/plain': TextIcon,
	'application/json': StructIcon,
	'application/x-yaml': StructIcon,
	'application/x-symlink': SymlinkIcon,
	'application/x-case': CaseIcon,
	'application/x-template': TemplateIcon,
	'application/x-list': ListIcon,
	'application/x-server': ListIcon,
	'application/x-server2': ListIcon,
};

interface ParamIconProps extends SvgIconProps {
	type: ParamType;
	path?: string;
	numChildren?: number;
}

export default function ParamIcon(props: ParamIconProps) {
	const { type, path, name, numChildren, ...rest } = props;
	let Icon = iconByType[type] || DefaultIcon;

	if (type === 'application/x-null') {
		if (path === '/') {
			Icon = RootIcon;
		} else if (path?.split('/').length === 2) {
			Icon = TopLevelIcon;
		} else if (numChildren !== undefined && numChildren > 0) {
			Icon = FolderIcon;
		}
	}
	return <Icon {...rest}/>;
}
