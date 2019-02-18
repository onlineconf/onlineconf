import * as React from 'react';
import { Link, LinkProps, Route } from 'react-router-dom';
import Drawer from '@material-ui/core/Drawer';
import { SvgIconProps } from '@material-ui/core/SvgIcon';
import { List, ListItem, ListItemIcon, ListItemText, Theme, withStyles, WithStyles, createStyles, withWidth, Divider, Typography } from '@material-ui/core';
import { isWidthUp } from '@material-ui/core/withWidth';
import { Breakpoint } from '@material-ui/core/styles/createBreakpoints';

import SettingsIcon from '@material-ui/icons/Settings';
import HistoryIcon from '@material-ui/icons/History';
import GroupIcon from '@material-ui/icons/Group';
import StorageIcon from '@material-ui/icons/Storage';

interface ListLinkProps extends LinkProps {
	Icon: React.ComponentType<SvgIconProps>;
	children: string;
	to: string;
}

class ListLink extends React.Component<ListLinkProps> {

	private renderLink = (props: LinkProps) => (
		<Link to={this.props.to} {...props}/>
	);

	render() {
		const { Icon, to, children } = this.props;
		return (
			<Route path={to} exact>
				{({ match }) => (
					<ListItem button component={this.renderLink} {...{ selected: match !== null }}>
						<ListItemIcon><Icon/></ListItemIcon>
						<ListItemText>{children}</ListItemText>
					</ListItem>
				)}
			</Route>
		);
	}

}

const ClassicLink = (props: LinkProps) => <a href="/classic/" {...props}/>;

const styles = (theme: Theme) => createStyles({
	paper: {
		zIndex: theme.zIndex.appBar - 1,
	},
	toolbar: {
		...theme.mixins.toolbar,
		paddingLeft: 2 * theme.spacing.unit,
		paddingRight: 2 * theme.spacing.unit,
		display: 'flex',
		alignItems: 'center',
	},
	classicIcon: {
		margin: '4px 20px 4px 4px',
	},
	classicText: {
		fontStyle: 'italic',
	},
});

interface LeftMenuProps {
	className?: string;
	open: boolean;
	onClose: () => void;
	width: Breakpoint;
}

const LeftMenu = (props: LeftMenuProps & WithStyles<typeof styles>) => {
	const { classes, ...rest } = props;
	const { toolbar: toolbarClassName, classicIcon, classicText, ...restClasses } = classes;
	return (
		<Drawer {...rest} classes={restClasses} variant={isWidthUp('sm', props.width) ? 'persistent' : 'temporary'}>
			<div className={toolbarClassName}>
				<Typography variant="h6" color="primary">OnlineConf</Typography>
			</div>
			<Divider/>
			<List component="nav">
				<ListLink to="/" Icon={SettingsIcon}>Configuration</ListLink>
				<ListLink to="/history/" Icon={HistoryIcon}>History</ListLink>
				<ListLink to="/server/" Icon={StorageIcon}>Servers</ListLink>
				<ListLink to="/access-group/" Icon={GroupIcon}>Access</ListLink>
				<ListItem button component={ClassicLink}>
					<img src="/classic/css/type/default.png" className={classicIcon}/>
					<ListItemText className={classicText}>Classic</ListItemText>
				</ListItem>
			</List>
		</Drawer>
	);
};

export default withStyles(styles)(withWidth()(LeftMenu));
