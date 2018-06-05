import * as React from 'react';
import scrollIntoView from 'scroll-into-view-if-needed';
import { Theme, withStyles, WithStyles, TextField, createStyles, MenuItem, Popper, MenuList, Paper, ClickAwayListener, PropTypes } from '@material-ui/core';

import { getParam } from '../api';

const styles = (theme: Theme) => createStyles({
	popper: {
		zIndex: theme.zIndex.modal + 1,
	},
	paper: {
		maxHeight: 48 * 8 + 2 * theme.spacing.unit,
		overflowY: 'scroll',
	},
});

interface Option {
	label: string;
	value: string;
}

interface PathFieldProps {
	label: string;
	value: string;
	onChange: (props: { target: { value: string } }) => void;
	symlink?: 'resolve' | 'follow';
	className?: string;
	required?: boolean;
	autoFocus?: boolean;
	fullWidth?: boolean;
	margin?: PropTypes.Margin;
}

interface PathFieldState {
	options: Option[];
	selected: string;
	menuWidth?: number;
}

class PathField extends React.Component<PathFieldProps & WithStyles<typeof styles>, PathFieldState> {

	state: PathFieldState = {
		options: [],
		selected: '',
	};

	cache: { [K: string]: Option[] } = {};

	inputRef = React.createRef<HTMLElement>();
	menuListRef = React.createRef<any>();

	componentDidUpdate() {
		if (this.menuListRef.current) {
			scrollIntoView(this.menuListRef.current.selectedItemRef, { scrollMode: 'if-needed' });
		}
	}

	componentDidMount() {
		window.addEventListener('resize', this.handleWindowResize);
	}

	componentWillUnmount() {
		window.removeEventListener('resize', this.handleWindowResize);
	}

	handleWindowResize = () => {
		if (this.inputRef.current) {
			this.setState({ menuWidth: this.inputRef.current.clientWidth });
		}
	}

	async loadOptions(value: string) {
		const prefix = value.replace(/\/[^\/]*$/, '');
		value = prefix === '' ? '/' : prefix;
		const cached = this.cache[value];
		if (cached !== undefined) {
			return cached;
		}
		const parent = await getParam(value, this.props.symlink);
		if (parent.children === undefined) {
			this.cache[value] = [];
			return [];
		}
		const options = parent.children.map(child => ({ label: child.name, value: prefix + '/' + child.name }));
		this.cache[value] = options;
		return options;
	}

	handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
		this.props.onChange(event);
		const value = event.target.value;
		const options = (await this.loadOptions(value))
			.filter(option => option.value.startsWith(value));
		this.setState(({ selected }) => (
			options.filter(o => o.value === selected).length > 0 ? { options, selected }
				: { options, selected: options.length > 0 ? options[0].value : '' }
		));
	}

	handleInputKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
		if (this.state.options.length === 0) {
			return;
		}
		switch (event.key) {
			case 'ArrowDown': {
				event.preventDefault();
				this.setState(({ selected, options }) => {
					for (let i = 0; i < options.length; i++) {
						if (options[i].value === selected) {
							if (i + 1 === options.length) {
								return null;
							}
							return { selected: options[i + 1].value };
						}
					}
					return { selected: options[0].value };
				});
				break;
			}
			case 'ArrowUp': {
				event.preventDefault();
				this.setState(({ selected, options }) => {
					for (let i = 0; i < options.length; i++) {
						if (options[i].value === selected) {
							if (i === 0) {
								return null;
							}
							return { selected: options[i - 1].value };
						}
					}
					return { selected: options[options.length - 1].value };
				});
				break;
			}
			case 'Enter': {
				event.preventDefault();
				this.setValue(this.state.selected);
				break;
			}
			case 'Escape': {
				event.preventDefault();
				this.closeMenu();
				break;
			}
		}
	}

	setValue(value: string) {
		this.props.onChange({ target: { value: value } });
		this.closeMenu();
		this.inputRef.current!.focus();
	}

	closeMenu() {
		this.setState({ options: [] });
	}

	render() {
		return (
			<React.Fragment>
				<TextField
					label={this.props.label}
					value={this.props.value}
					onChange={this.handleChange}
					inputRef={this.inputRef}
					onKeyDown={this.handleInputKeyDown}
					required={this.props.required}
					autoFocus={this.props.autoFocus}
					margin={this.props.margin}
					fullWidth={this.props.fullWidth}
					className={this.props.className}
				/>
				<Popper
					open={this.state.options.length > 0}
					anchorEl={this.inputRef.current}
					className={this.props.classes.popper}
					placement="bottom-start"
					modifiers={{ flip: { enabled: true }, preventOverflow: { enabled: true, boundariesElement: 'viewport' } }}
				>
					<Paper
						elevation={8}
						className={this.props.classes.paper}
						style={this.inputRef.current !== null ? { width: this.inputRef.current!.clientWidth } : undefined}
					>
						<ClickAwayListener onClickAway={() => this.closeMenu()}>
							<MenuList { ...{ ref: this.menuListRef } }>
								{this.state.options.map(option => (
									<MenuItem
										onClick={() => this.setValue(option.value)}
										selected={option.value === this.state.selected}
									>
										{option.label}
									</MenuItem>
								))}
							</MenuList>
						</ClickAwayListener>
					</Paper>
				</Popper>
			</React.Fragment>
		);
	}

}

export default withStyles(styles)(PathField);
