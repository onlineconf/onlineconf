
import * as React from 'react';
import md5 from 'crypto-js/md5';
import { Avatar as MUIAvatar, AvatarProps as MUIAvatarProps, makeStyles, Tooltip } from '@material-ui/core';

import { AvatarContext } from './UIConfig';

const useStyles = makeStyles({
	link: {
		textDecoration: 'none',
	}
});

type AvatarProps = Omit<MUIAvatarProps, 'children' | 'src'> & {
	username: string;
	disableTooltip?: boolean;
};

export default function Avatar(props: AvatarProps) {
	const { username, disableTooltip, ...rest } = props;
	const classes = useStyles();
	const config = React.useContext(AvatarContext);
	const path = React.useMemo(() => {
		if (config === undefined) {
			return;
		}
		let path = (config.rename && config.rename[username]) || username;
		if (config.domain) {
			path += '@' + config.domain;
		}
		if (config.gravatar) {
			path = md5(path.toLowerCase()).toString();
		}
		return `${config.uri}/${path}`;
	}, [username, config]);
	const letters = React.useMemo(() => {
		const m = /^(?:(\w)[^.]*?\.(\w)|(\w{2}))/.exec(username);
		return m?.slice(1).join('').toUpperCase();
	}, [username]);
	let avatar = <MUIAvatar {...rest} src={path}>{letters}</MUIAvatar>;
	if (config?.link) {
		const linkUsername = (config.link.rename && config.link.rename[username]) || username;
		avatar = <a href={`${config.link.uri}/${linkUsername}`} target="_blank" rel="noreferrer" className={classes.link}>{avatar}</a>;
	}
	if (!disableTooltip) {
		avatar = <Tooltip title={username} placement="top" arrow>{avatar}</Tooltip>;
	}
	return avatar;
}
