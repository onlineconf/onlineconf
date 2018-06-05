import * as React from 'react';
import { Link, LinkProps, Route } from 'react-router-dom';
import Drawer from '@material-ui/core/Drawer';
import { SvgIconProps } from '@material-ui/core/SvgIcon';
import { List, ListItem, ListItemIcon, ListItemText, Theme, withStyles, WithStyles } from '@material-ui/core';

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

const styles = (theme: Theme) => ({
	paper: {
		zIndex: theme.zIndex.appBar - 1,
	},
	toolbar: theme.mixins.toolbar,
});

interface LeftMenuProps {
	className?: string;
	open: boolean;
	onClose: () => void;
}

const LeftMenu = (props: LeftMenuProps & WithStyles<typeof styles>) => {
	const { classes, ...rest } = props;
	const { toolbar: toolbarClassName, ...restClasses } = classes;
	return (
		<Drawer {...rest} classes={restClasses} variant="persistent">
			<div className={toolbarClassName}/>
			<List component="nav">
				<ListLink to="/" Icon={SettingsIcon}>Configuration</ListLink>
				<ListLink to="/history/" Icon={HistoryIcon}>History</ListLink>
				<ListLink to="/server/" Icon={StorageIcon}>Servers</ListLink>
				<ListLink to="/access/" Icon={GroupIcon}>Access</ListLink>
			</List>
		</Drawer>
	);
};

export default withStyles(styles)(LeftMenu);
