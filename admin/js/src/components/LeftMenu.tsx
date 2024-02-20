import * as React from 'react';
import { Link, LinkProps, Route } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Theme, useTheme, alpha } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import Drawer from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';
import useMediaQuery from '@mui/material/useMediaQuery';
import { SvgIconProps } from '@mui/material/SvgIcon';
import ButtonGroup from '@mui/material/ButtonGroup';
import Button, { ButtonProps } from '@mui/material/Button';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import InputAdornment from '@mui/material/InputAdornment';

import SettingsIcon from '@mui/icons-material/Settings';
import HistoryIcon from '@mui/icons-material/History';
import GroupIcon from '@mui/icons-material/Group';
import StorageIcon from '@mui/icons-material/Storage';
import LanguageIcon from '@mui/icons-material/Translate';
import LightModeIcon from '@mui/icons-material/Brightness7';
import DarkModeIcon from '@mui/icons-material/Brightness4';
import SystemModeIcon from '@mui/icons-material/SettingsBrightness';

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
				{({ match }: { match: unknown }) => (
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
			backgroundColor: alpha(theme.palette.action.selected, theme.palette.action.selectedOpacity + theme.palette.action.hoverOpacity),
		}
	},
}), { name: 'LeftMenu' });

interface LeftMenuProps {
	className?: string;
	open: boolean;
	onClose: () => void;
	paletteType: PaletteType;
	onChangePaletteType(value: PaletteType): void;
	onSlideEnd?: () => void;
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
	const variant = useMediaQuery(theme.breakpoints.up('md')) ? 'persistent' : 'temporary';
	return (
		<Drawer
			open={props.open}
			onClose={props.onClose}
			variant={variant}
			classes={{ paper: classes.paper }}
			SlideProps={{ onEntered: props.onSlideEnd, onExited: props.onSlideEnd }}
		>
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
