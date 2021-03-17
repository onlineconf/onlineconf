import * as React from 'react';
import { Link, LinkProps, Route } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Theme, useTheme, makeStyles, fade } from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Divider from '@material-ui/core/Divider';
import Typography from '@material-ui/core/Typography';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import { SvgIconProps } from '@material-ui/core/SvgIcon';
import ButtonGroup from '@material-ui/core/ButtonGroup';
import Button, { ButtonProps } from '@material-ui/core/Button';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import InputAdornment from '@material-ui/core/InputAdornment';

import SettingsIcon from '@material-ui/icons/Settings';
import HistoryIcon from '@material-ui/icons/History';
import GroupIcon from '@material-ui/icons/Group';
import StorageIcon from '@material-ui/icons/Storage';
import LanguageIcon from '@material-ui/icons/Translate';
import LightModeIcon from '@material-ui/icons/Brightness7';
import DarkModeIcon from '@material-ui/icons/Brightness4';
import SystemModeIcon from '@material-ui/icons/SettingsBrightness';

import { PaletteType } from './UIConfig';

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
	fill: {
		height: '100%',
	},
	selected: {
		backgroundColor: theme.palette.action.selected,
		'&:hover': {
			backgroundColor: fade(theme.palette.action.selected, theme.palette.action.selectedOpacity + theme.palette.action.hoverOpacity),
		}
	},
}), { name: 'LeftMenu' });

interface LeftMenuProps {
	className?: string;
	open: boolean;
	onClose: () => void;
	paletteType: PaletteType;
	onChangePaletteType(value: PaletteType): void;
}

export default function LeftMenu(props: LeftMenuProps) {
	const classes = useStyles();
	const theme = useTheme();
	const { t, i18n } = useTranslation();
	const themeButtonOpts = (paletteType: PaletteType): ButtonProps => ({
		variant: 'outlined',
		className: props.paletteType === paletteType ? classes.selected : undefined,
		onClick: () => props.onChangePaletteType(paletteType),
	});
	const variant = useMediaQuery(theme.breakpoints.up('sm')) ? 'persistent' : 'temporary';
	return (
		<Drawer open={props.open} onClose={props.onClose} classes={{ paper: classes.paper }} variant={variant}>
			<div className={classes.toolbar}>
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
						<img src="/classic/css/type/default.png" className={classes.classicIcon} alt=""/>
					</ListItemIcon>
					<ListItemText className={classes.classicText}>Classic</ListItemText>
				</ListItem>
			</List>
			<div className={classes.fill}/>
			<List component="div">
				<ListItem component="div">
					<Select
						value={i18n.language.split('-', 1)[0]}
						onChange={event => i18n.changeLanguage(event.target.value as string)}
						variant="outlined"
						fullWidth
						margin="dense"
						startAdornment={<InputAdornment position="start"><LanguageIcon/></InputAdornment>}
					>
						<MenuItem value="en">English</MenuItem>
						<MenuItem value="ru">Русский</MenuItem>
					</Select>
				</ListItem>
				<ListItem component="div">
					<ButtonGroup fullWidth>
						<Button {...themeButtonOpts('light')}><LightModeIcon/></Button>
						<Button {...themeButtonOpts('system')}><SystemModeIcon/></Button>
						<Button {...themeButtonOpts('dark')}><DarkModeIcon/></Button>
					</ButtonGroup>
				</ListItem>
			</List>
		</Drawer>
	);
}
