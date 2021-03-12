import * as React from 'react';
import { SvgIconProps } from '@material-ui/core/SvgIcon';

import DefaultIcon from '@material-ui/icons/Settings';
import RootIcon from '@material-ui/icons/Language';
import TopLevelIcon from '@material-ui/icons/Domain';
import NullIcon from '@material-ui/icons/Block';
import FolderIcon from '@material-ui/icons/FolderOpen';
import TextIcon from '@material-ui/icons/Subject';
import StructIcon from '@material-ui/icons/DeviceHub';
import SymlinkIcon from '@material-ui/icons/Launch';
import CaseIcon from '@material-ui/icons/Help';
import TemplateIcon from '@material-ui/icons/LocalAtm';
import ListIcon from '@material-ui/icons/List';

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
