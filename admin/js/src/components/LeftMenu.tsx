import * as React from 'react';
import { Link, LinkProps, Route } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Theme, useTheme, makeStyles } from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Divider from '@material-ui/core/Divider';
import Typography from '@material-ui/core/Typography';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import { SvgIconProps } from '@material-ui/core/SvgIcon';

import SettingsIcon from '@material-ui/icons/Settings';
import HistoryIcon from '@material-ui/icons/History';
import GroupIcon from '@material-ui/icons/Group';
import StorageIcon from '@material-ui/icons/Storage';

export const leftMenuWidth = 240;

interface ListLinkProps extends LinkProps {
	Icon: React.ComponentType<SvgIconProps>;
	children: string;
	to: string;
}

class ListLink extends React.Component<ListLinkProps> {
	render() {
		const { Icon, to, children } = this.props;
		return (
			<Route path={to} exact>
				{({ match }) => (
					<ListItem button component={Link} to={to} selected={ match !== null }>
						<ListItemIcon><Icon/></ListItemIcon>
						<ListItemText>{children}</ListItemText>
					</ListItem>
				)}
			</Route>
		);
	}
}

const ClassicLink = React.forwardRef(
	function ClassicLink(props: LinkProps, ref: React.Ref<HTMLAnchorElement>) {
		// eslint-disable-next-line jsx-a11y/anchor-has-content
		return <a href="/classic/" {...props} ref={ref}/>;
	}
);

const useStyles = makeStyles((theme: Theme) => ({
	paper: {
		width: leftMenuWidth,
		zIndex: theme.zIndex.appBar - 1,
	},
	toolbar: {
		...theme.mixins.toolbar,
		paddingLeft: theme.spacing(2),
		paddingRight: theme.spacing(2),
		display: 'flex',
		alignItems: 'center',
	},
	classicIcon: {
		margin: 4,
		width: 16,
		height: 16,
	},
	classicText: {
		fontStyle: 'italic',
	},
}));

interface LeftMenuProps {
	className?: string;
	open: boolean;
	onClose: () => void;
}

export default function LeftMenu(props: LeftMenuProps) {
	const classes = useStyles();
	const { toolbar: toolbarClassName, classicIcon, classicText, ...restClasses } = classes;
	const theme = useTheme();
	const { t } = useTranslation();
	return (
		<Drawer {...props} classes={restClasses} variant={useMediaQuery(theme.breakpoints.up('sm')) ? 'persistent' : 'temporary'}>
			<div className={toolbarClassName}>
				<Typography variant="h6" color="primary">OnlineConf</Typography>
			</div>
			<Divider/>
			<List component="nav">
				<ListLink to="/" Icon={SettingsIcon}>{t('left.configuration')}</ListLink>
				<ListLink to="/history/" Icon={HistoryIcon}>{t('left.history')}</ListLink>
				<ListLink to="/server/" Icon={StorageIcon}>{t('left.servers')}</ListLink>
				<ListLink to="/access-group/" Icon={GroupIcon}>{t('left.access')}</ListLink>
				<ListItem button component={ClassicLink} to="/classic/">
					<ListItemIcon>
						<img src="/classic/css/type/default.png" className={classicIcon} alt=""/>
					</ListItemIcon>
					<ListItemText className={classicText}>Classic</ListItemText>
				</ListItem>
			</List>
		</Drawer>
	);
}
