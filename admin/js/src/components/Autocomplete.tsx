import * as React from 'react';
import scrollIntoView from 'scroll-into-view-if-needed';
import { Theme, withStyles, WithStyles, TextField, createStyles, MenuItem, Popper, MenuList, Paper, ClickAwayListener, Omit } from '@material-ui/core';
import { BaseTextFieldProps } from '@material-ui/core/TextField';

const styles = (theme: Theme) => createStyles({
	popper: {
		zIndex: theme.zIndex.modal + 1,
	},
	paper: {
		maxHeight: 48 * 8 + 2 * theme.spacing.unit,
		overflowY: 'scroll',
		scrollBehavior: 'smooth',
	},
});

export interface AutocompleteOption {
	label: string;
	value: string;
}

export interface AutocompleteProps extends Omit<BaseTextFieldProps, 'classes' | 'onError'> {
	value: string;
	onChange: (props: { target: { value: string } }) => void;
	variant?: 'standard' | 'outlined' | 'filled';
	loadOptions: (value: string) => Promise<AutocompleteOption[]>;
}

interface AutocompleteState {
	options: AutocompleteOption[];
	selected: string;
	menuWidth?: number;
}

class PathField extends React.Component<AutocompleteProps & WithStyles<typeof styles>, AutocompleteState> {

	state: AutocompleteState = {
		options: [],
		selected: '',
	};

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

	handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
		this.props.onChange(event);
		const value = event.target.value;
		const options = await this.props.loadOptions(value);
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
		this.props.onChange({ target: { value } });
		this.closeMenu();
		this.inputRef.current!.focus();
	}

	closeMenu() {
		this.setState({ options: [] });
	}

	render() {
		const { onChange, loadOptions, classes, ...props } = this.props;
		return (
			<React.Fragment>
				<TextField
					{...props}
					variant={this.props.variant as any}
					onChange={this.handleChange}
					inputRef={this.inputRef}
					onKeyDown={this.handleInputKeyDown}
				/>
				<Popper
					open={this.state.options.length > 0}
					anchorEl={this.inputRef.current}
					className={classes.popper}
					placement="bottom-start"
				>
					<Paper
						elevation={8}
						className={classes.paper}
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
