import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme, makeStyles, useTheme } from '@material-ui/core/styles';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import TextField from '@material-ui/core/TextField';
import InputAdornment from '@material-ui/core/InputAdornment';
import IconButton from '@material-ui/core/IconButton';
import useMediaQuery from '@material-ui/core/useMediaQuery';

import MenuIcon from '@material-ui/icons/Menu';
import SearchIcon from '@material-ui/icons/Search';
import IconButtonProgress from './IconButtonProgress';

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
	}

	handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		this.props.onSearch(this.state.value);
	}

	render() {
		return (
			<form onSubmit={this.handleSubmit}>
				<TextField
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
									<IconButton type="submit">
										<SearchIcon />
									</IconButton>
								</IconButtonProgress>
							</InputAdornment>
						)
					}}
				/>
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
	username?: string;
	onMenu: () => void;
	onSearch: (term: string) => void;
	searching: boolean;
}

export default function TopBar(props: TopBarProps) {
	const classes = useStyles();
	const theme = useTheme();
	const showAvatar = useMediaQuery(theme.breakpoints.up('sm'));
	return (
		<AppBar position="sticky" color="default" className={classes.root}>
			<Toolbar>
				<IconButton color="inherit" className={classes.menuButton} onClick={props.onMenu}>
					<MenuIcon/>
				</IconButton>
				<Typography variant="h6" color="inherit" className={classes.title}>
					OnlineConf
				</Typography>
				<Search onSearch={props.onSearch} searching={props.searching} className={classes.search}/>
				{showAvatar && props.username && <Avatar username={props.username} className={classes.avatar}/>}
			</Toolbar>
		</AppBar>
	);
}
