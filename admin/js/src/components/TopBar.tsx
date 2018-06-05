import * as React from 'react';
import AppBar from '@material-ui/core/AppBar';
import { WithStyles, withStyles, Theme } from '@material-ui/core/styles';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import TextField from '@material-ui/core/TextField';
import InputAdornment from '@material-ui/core/InputAdornment';
import IconButton from '@material-ui/core/IconButton';

import MenuIcon from '@material-ui/icons/Menu';
import SearchIcon from '@material-ui/icons/Search';
import IconButtonProgress from './IconButtonProgress';

interface SearchProps {
	onSearch: (term: string) => void;
	searching: boolean;
	className?: string;
}

interface SearchState {
	value: string;
}

class Search extends React.Component<SearchProps, SearchState> {

	state = {
		value: '',
	};

	handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		if (event.target.value === '' && this.state.value !== '') {
			this.props.onSearch('');
		}
		this.setState({ value: event.target.value });
	}

	handleSubmit = (event: React.FormEvent<{}>) => {
		event.preventDefault();
		this.props.onSearch(this.state.value);
	}

	render() {
		return (
			<TextField
				placeholder="Search"
				type="search"
				margin="none"
				color="inherit"
				value={this.state.value}
				onChange={this.handleChange}
				onSubmit={this.handleSubmit}
				component="form"
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
		);
	}

}

const styles = (theme: Theme) => ({
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
		marginLeft: 2 * theme.spacing.unit,
	}
});

interface TopBarProps {
	onMenu: () => void;
	onSearch: (term: string) => void;
	searching: boolean;
}

const TopBar = (props: TopBarProps & WithStyles<typeof styles>) => (
	<AppBar position="sticky" color="default" className={props.classes.root}>
		<Toolbar>
			<IconButton color="inherit" className={props.classes.menuButton} onClick={props.onMenu}>
				<MenuIcon/>
			</IconButton>
			<Typography variant="title" color="inherit" className={props.classes.title}>
				OnlineConf
			</Typography>
			<Search onSearch={props.onSearch} searching={props.searching} className={props.classes.search}/>
		</Toolbar>
	</AppBar>
);

export default withStyles(styles)(TopBar);
