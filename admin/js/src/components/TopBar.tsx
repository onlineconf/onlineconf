import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme, useTheme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import useMediaQuery from '@mui/material/useMediaQuery';

import MenuIcon from '@mui/icons-material/Menu';
import SearchIcon from '@mui/icons-material/Search';
import IconButtonProgress from './IconButtonProgress';

import WhoAmIContext from './WhoAmIContext';
import Avatar from './Avatar';

interface SearchProps {
	onSearch: (term: string) => void;
	searching: boolean;
	className?: string;
}

interface SearchState {
	value: string;
}

class SearchBase extends React.Component<SearchProps & WithTranslation, SearchState> {

	state = {
		value: '',
	};

	handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		if (event.target.value === '' && this.state.value !== '') {
			this.props.onSearch('');
		}
		this.setState({ value: event.target.value });
	};

	handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		this.props.onSearch(this.state.value);
	};

	render() {
		return (
			<form onSubmit={this.handleSubmit}>
				<TextField
					variant="standard"
					placeholder={this.props.t('search')}
					type="search"
					margin="none"
					value={this.state.value}
					onChange={this.handleChange}
					className={this.props.className}
					InputProps={{
						endAdornment: (
							<InputAdornment position="end">
								<IconButtonProgress loading={this.props.searching}>
									<IconButton type="submit" size="large">
										<SearchIcon />
									</IconButton>
								</IconButtonProgress>
							</InputAdornment>
						)
					}} />
			</form>
		);
	}

}

const Search = withTranslation()(SearchBase);

const useStyles = makeStyles((theme: Theme) => ({
	root: {
	},
	title: {
		flex: 1,
	},
	menuButton: {
		marginLeft: -12,
		marginRight: 4,
	},
	search: {
		marginLeft: theme.spacing(2),
	},
	avatar: {
		width: 32,
		height: 32,
		marginLeft: 12,
	}
}));

interface TopBarProps {
	onMenu: () => void;
	onSearch: (term: string) => void;
	searching: boolean;
}

export default function TopBar(props: TopBarProps) {
	const classes = useStyles();
	const theme = useTheme();
	const showAvatar = useMediaQuery(theme.breakpoints.up('sm'));
	const { username } = React.useContext(WhoAmIContext);
	return (
		<AppBar position="sticky" color="default" className={classes.root}>
			<Toolbar>
				<IconButton
					color="inherit"
					className={classes.menuButton}
					onClick={props.onMenu}
					size="large">
					<MenuIcon/>
				</IconButton>
				<Typography variant="h6" color="inherit" className={classes.title}>
					OnlineConf
				</Typography>
				<Search onSearch={props.onSearch} searching={props.searching} className={classes.search}/>
				{showAvatar && username && <Avatar username={username} className={classes.avatar}/>}
			</Toolbar>
		</AppBar>
	);
}
