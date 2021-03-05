import * as React from 'react';
import { History } from 'history';

import { Theme, withStyles, WithStyles } from '@material-ui/core/styles';

import * as API from '../api';
import { IParamNode } from './common';
import ConfigTreeNode from './ConfigTreeNode';
import Editor from './Editor';
import ParamAccess from './ParamAccess';
import ParamDeleteConfirm from './ParamDeleteConfirm';
import ParamDescribe from './ParamDescribe';
import ParamLog from './ParamLog';
import ParamMove from './ParamMove';
import ParamNotification from './ParamNotification';
import Tree from './Tree';
import ValuePopover from './ValuePopover';
import Viewer from './Viewer';
import { TreeNodeState } from './TreeNode';

const styles = (theme: Theme) => ({
	icon: {
		'color': '#888888',
		'font-size': '18px',
		'vertical-align': 'middle',
	},
});

function paramNode(param: API.IParam, options: { open?: boolean, selected?: string } = {}): IParamNode {
	const { children, ...rest } = param;
	const state = options.open && children ? 'open' : param.num_children > 0 ? 'closed' : 'leaf';
	const node: IParamNode = { ...rest, state };

	if (children) {
		node.children = children.reduce((obj, param) => {
			obj[param.name] = paramNode(param, options);
			return obj;
		}, {} as { [k: string] : IParamNode });
	}

	if (node.path === options.selected) {
		node.selected = true;
	}

	return node;
}

function updateTree(root: IParamNode | undefined, param: API.IParam, options: { open?: boolean, selected?: string } = {}): IParamNode | undefined {
	if (param.path === '/') {
		return paramNode(param, options);
	}

	let parent = root;
	const chunks = param.path.split('/').slice(1, -1);

	for (const chunk of chunks) {
		if (!(parent && parent.children)) {
			return root;
		}
		parent = parent.children[chunk];
	}

	if (!(parent && parent.children)) {
		return root;
	}
	parent.children[param.name] = paramNode(param, options);

	return root;
}

function getNodeAndParents(root: IParamNode, path: string): { node: IParamNode, parents: IParamNode[] } | undefined {
	let node = root;
	const parents = [];

	if (path !== '/') {
		const chunks = path.split('/').slice(1);
		for (const chunk of chunks) {
			if (!(node && node.children)) {
				return;
			}
			parents.push(node);
			node = node.children[chunk];
		}
	}

	if (!node) {
		return;
	}

	return { node, parents };
}

function getNode(root: IParamNode, path: string): IParamNode | undefined {
	let node = root;

	if (path !== '/') {
		const chunks = path.split('/').slice(1);
		for (const chunk of chunks) {
			if (!(node && node.children)) {
				return;
			}
			node = node.children[chunk];
		}
	}

	return node;
}

function modifyNode(root: IParamNode, path: string, action: (node: IParamNode) => void): IParamNode {
	const node = getNode(root, path);

	if (node) {
		action(node);
	}

	return root;
}

function modifyNodes(root: IParamNode, paths: string[], action: (node: IParamNode) => void): IParamNode {
	for (const path of paths) {
		root = modifyNode(root, path, action);
	}
	return root;
}

function setNodeState(root: IParamNode, path: string, state: TreeNodeState = 'open'): IParamNode {
	const node = getNode(root, path);

	if (node) {
		node.state = state;
	}

	return root;
}

function initMatch(param: IParamNode) {
	param.hidden = true;
	param.match = false;
	param.forceOpen = false;
	if (param.children) {
		for (const name of Object.keys(param.children)) {
			initMatch(param.children[name]);
		}
	}
}

function clearMatch(param: IParamNode) {
	delete param.hidden;
	delete param.match;
	delete param.forceOpen;
	if (param.children) {
		for (const name of Object.keys(param.children)) {
			clearMatch(param.children[name]);
		}
	}
}

function setMatch(root: IParamNode, paths: string[]) {
	initMatch(root);
	for (const path of paths) {
		const ret = getNodeAndParents(root, path);
		if (ret) {
			ret.node.match = true;
			ret.node.hidden = false;
			for (const param of ret.parents) {
				param.hidden = false;
				param.forceOpen = true;
			}
		}
	}
}

function showSelected(root: IParamNode, selected?: string) {
	if (!selected) {
		return;
	}

	const ret = getNodeAndParents(root, selected);
	if (!ret) {
		return;
	}

	for (const param of ret.parents) {
		if (param.state === 'closed') {
			param.state = 'open';
		}
	}
}

function parentPath(path: string) {
	// eslint-disable-next-line no-useless-escape
	let parentPath = path.replace(/\/[^\/]+$/, '');
	if (parentPath === '') {
		parentPath = '/';
	}
	return parentPath;
}

interface ConfigTreeProps {
	history: History;
	search: string;
	onSearching: (searching: boolean) => void;
	onError: (error: Error) => void;
}

interface ConfigTreeState {
	root?: IParamNode;
	selected?: string;
	menu?: string;
	menuAnchorX?: number;
	dialog: JSX.Element | null;
	popover: JSX.Element | null;
}

class ConfigTree extends React.Component< ConfigTreeProps & WithStyles<'icon'>, ConfigTreeState > {

	public state: ConfigTreeState = {
		popover: null,
		dialog: null,
	};

	historyUnlisten?: () => void;

	componentDidMount() {
		const { history } = this.props;

		this.historyUnlisten = history.listen(location => {
			const path = location.hash.substring(1);
			if (this.state.selected !== path) {
				this.selectNode(path);
			}
		});

		const path = history.location.hash.substring(1);
		if (path) {
			this.selectNode(path);
		} else {
			this.loadNode('/', { open: true });
		}
	}

	componentWillUnmount() {
		this.historyUnlisten!();
	}

	componentDidUpdate(prevProps: ConfigTreeProps) {
		if (prevProps.search !== this.props.search) {
			if (this.props.search) {
				this.search(this.props.search);
			} else {
				this.clearSearch();
			}
		}
	}

	requiredPaths(path: string, withChildren = false) {
		const paths: string[] = [];

		let node = this.state.root;
		if (!node) {
			paths.push('/');
		}
		let currentPath = '';
		const chunks = path.split('/').slice(1);
		for (const chunk of chunks) {
			if (node && node.children) {
				node = node.children[chunk];
			} else if (currentPath) {
				paths.push(currentPath);
			}
			currentPath += '/' + chunk;
		}
		if (withChildren && !(node && node.children)) {
			paths.push(currentPath);
		}

		return paths;
	}

	async loadNodes(paths: string[], options: { open?: boolean, withChildren?: boolean, reload?: boolean } = {}) {
		try {
			const loadPaths = new Set<string>(options.reload ? paths : []);
			for (const path of paths) {
				for (const p of this.requiredPaths(path, options.withChildren)) {
					loadPaths.add(p);
				}
			}
			this.setState(({ root }) => ({ root: modifyNodes(root!, Array.from(loadPaths), node => { node.state = 'loading'; }) }));
			const nodes = await API.getParams(loadPaths);

			this.setState(prevState => {
				let root = prevState.root;
				for (const path of Object.keys(nodes).sort()) {
					root = updateTree(root, nodes[path], { open: options.open, selected: prevState.selected });
				}
				return { root };
			});
		} catch (error) {
			this.setState(({ root }) => ({ root: modifyNodes(root!, paths, node => { node.state = 'closed'; }) }));
			throw error;
		}
	}

	async loadNode(path: string, options?: { open?: boolean, withChildren?: boolean, reload?: boolean }) {
		try {
			await this.loadNodes([path], options);
		} catch (error) {
			this.props.onError(error);
		}
	}

	selectNode(path: string) {
		let shouldLoad = false;
		this.setState(prevState => {
			let treeModified = false;

			if (prevState.selected) {
				if (prevState.selected === path) {
					return null;
				}

				const node = getNode(prevState.root!, prevState.selected);
				if (node) {
					node.selected = false;
					treeModified = true;
				}
			}

			const ret = getNodeAndParents(prevState.root!, path);
			if (ret) {
				ret.node.selected = true;
				for (const parent of ret.parents) {
					if (parent.state === 'closed') {
						parent.state = 'open';
					}
				}
				treeModified = true;
			} else {
				shouldLoad = true;
			}

			if (treeModified) {
				return { selected: path, root: prevState.root };
			} else {
				return { selected: path };
			}
		}, () => {
			if (shouldLoad) {
				this.loadNode(path, { open: true });
			}
		});
	}

	async search(term: string) {
		const { onSearching, onError } = this.props;
		onSearching(true);
		try {
			const matchNodes = await API.search(term);
			const matchPaths = matchNodes.map(param => param.path);
			await this.loadNodes(matchPaths);
			this.setState(({ root }) => {
				setMatch(root!, matchPaths);
				return { root };
			});
		} catch (error) {
			onError(error);
		}
		onSearching(false);
	}

	clearSearch() {
		this.setState(({ root, selected }) => {
			clearMatch(root!);
			showSelected(root!, selected);
			return { root };
		});
	}

	handleOpen = (path: string) => {
		const node = getNode(this.state.root!, path);

		if (node && node.children) {
			this.setState({ root: setNodeState(this.state.root!, path, 'open') });
		} else {
			this.loadNode(path, { open: true, withChildren: true });
		}
	}

	handleClose = (path: string) => {
		this.setState({ root: setNodeState(this.state.root!, path, 'closed') });
	}

	handleSelect = (param: IParamNode) => {
		this.selectNode(param.path);
		this.props.history.push({ hash: '#' + param.path });
	}

	handleMenuOpen = (path: string, anchorX?: number) => {
		this.setState({ menu: path, menuAnchorX: anchorX });
	}

	handleMenuClose = (path: string) => {
		this.setState(({ menu }) => (menu === path ? { menu: undefined } : null));
	}

	handleReload = async (param: IParamNode) => {
		await this.loadNode(param.path, { reload: true, open: param.state === 'open' });
		this.setState(({ menu }) => (menu === param.path ? { menu: undefined } : null ));
	}

	handleView = (param: IParamNode) => {
		this.setState({
			dialog: <Viewer
				param={param}
				onClose={this.handleDialogClose}
				onEdit={() => this.handleEdit(param)}
				onDescribe={() => this.handleDescribe(param)}
			/>,
			menu: undefined,
		});
	}

	handleEdit = (param: IParamNode) => {
		this.setState({
			dialog: <Editor
				path={param.path}
				version={param.version}
				type={param.mime}
				value={param.data}
				summary={param.summary}
				description={param.description}
				onChange={this.handleEditDone}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleEditDone = (param: API.IParam) => {
		this.setState(({ root }) => ({
			root: modifyNode(root!, param.path, node => {
				Object.assign(node, paramNode(param));
			}),
		}));
	}

	handleAddChild = (param: IParamNode) => {
		this.setState({
			dialog: <Editor
				create
				path={param.path}
				notification={param.notification}
				onChange={this.handleAddChildDone}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleAddChildDone = (param: API.IParam) => {
		this.setState(({ root }) => {
			const parent = getNode(root!, parentPath(param.path));
			if (parent === undefined) {
				return null;
			}
			if (parent.children === undefined) {
				if (parent.num_children !== 0) {
					return null;
				}
				parent.children = {};
				parent.state = 'open';
			}
			if (parent.children[param.name] === undefined) {
				parent.num_children++;
			}
			parent.children[param.name] = paramNode(param);
			return { root };
		});
	}

	handleDelete = (param: IParamNode) => {
		this.setState({
			dialog: <ParamDeleteConfirm
				path={param.path}
				version={param.version}
				onDeleted={() => this.handleDeleted(param)}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleDeleted = (param: IParamNode) => {
		this.setState(({ root }) => {
			const parent = getNode(root!, parentPath(param.path));
			if (parent !== undefined && parent.children !== undefined) {
				if (parent.children[param.name] !== undefined) {
					parent.num_children--;
				}
				delete(parent.children[param.name]);
				return { root, dialog: null };
			}
			return { dialog: null };
		});
	}

	handleMove = (param: IParamNode) => {
		this.setState({
			dialog: <ParamMove
				path={param.path}
				version={param.version}
				onMoved={path => this.handleMoved(param, path)}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleMoved = async (param: IParamNode, path: string) => {
		try {
			await this.loadNodes([parentPath(param.path), parentPath(path)], {
				reload: true,
				open: true,
				withChildren: true,
			});
			this.selectNode(path);
			this.props.history.push({ hash: '#' + path });
			this.setState({ dialog: null });
		} catch (error) {
			this.props.onError(error);
		}
	}

	handleDescribe = (param: IParamNode) => {
		this.setState({
			dialog: <ParamDescribe
				path={param.path}
				summary={param.summary}
				description={param.description}
				onDescribed={this.handleDescribed}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleDescribed = (param: API.IParam) => {
		this.setState(({ root }) => ({
			root: modifyNode(root!, param.path, node => {
				Object.assign(node, paramNode(param));
			}),
			dialog: null,
		}));
	}

	handleNotification = (param: IParamNode) => {
		this.setState({
			dialog: <ParamNotification
				path={param.path}
				overridden={param.notification_modified}
				value={param.notification}
				onChange={this.handleModifyNotification}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
			menu: undefined,
		});
	}

	handleModifyNotification = (param: API.IParam) => {
		this.setState(({ root }) => ({
			root: modifyNode(root!, param.path, node => {
				node.notification = param.notification;
				node.notification_modified = param.notification_modified;
			}),
			dialog: <ParamNotification
				path={param.path}
				overridden={param.notification_modified}
				value={param.notification}
				onChange={this.handleModifyNotification}
				onError={this.props.onError}
				onClose={this.handleDialogClose}
			/>,
		}));
	}

	handleLog = (path: string) => {
		this.setState((prevState) => {
			const onLoaded = () => {
				this.setState(prevState => ({
					root: modifyNode(prevState.root!, path, node => { delete (node.logLoading); }),
					menu: undefined,
				}));
			};
			const onError = (error: Error) => {
				this.setState(prevState => ({
					root: modifyNode(prevState.root!, path, node => { delete(node.logLoading); }),
					dialog: null,
				}));
				this.props.onError(error);
			};
			return {
				root: modifyNode(prevState.root!, path, node => { node.logLoading = true; }),
				dialog: <ParamLog path={path} onClose={this.handleDialogClose} onLoaded={onLoaded} onError={onError} onChange={this.handleEditDone}/>,
			};
		});
	}

	// TODO apply changes to tree
	handleAccess = (path: string) => {
		this.setState((prevState) => {
			const onLoaded = () => {
				this.setState(prevState => ({
					root: modifyNode(prevState.root!, path, node => { delete (node.accessLoading); }),
					menu: undefined,
				}));
			};
			const onError = (error: Error) => {
				this.setState(prevState => ({
					root: modifyNode(prevState.root!, path, node => { delete(node.accessLoading); }),
					dialog: null,
				}));
				this.props.onError(error);
			};
			return {
				root: modifyNode(prevState.root!, path, node => { node.accessLoading = true; }),
				dialog: <ParamAccess path={path} onClose={this.handleDialogClose} onLoaded={onLoaded} onError={onError}/>,
			};
		});
	}

	handleDialogClose = () => {
		this.setState({ dialog: null });
	}

	handleValuePopoverOpen = (event: React.MouseEvent<HTMLElement>, param: IParamNode) => {
		this.setState({
			popover: <ValuePopover type={param.mime} value={param.data} onClose={this.handleValuePopoverClose} anchorEl={event.currentTarget as HTMLElement} />
		});
	}

	handleValuePopoverClose = () => {
		this.setState({ popover: null });
	}

	public render() {
		return (
			<div>
				<Tree>
					{this.state.root && (
						<ConfigTreeNode
							param={this.state.root}
							menu={this.state.menu}
							menuAnchorX={this.state.menuAnchorX}
							onOpen={this.handleOpen}
							onClose={this.handleClose}
							onSelect={this.handleSelect}
							onMenuOpen={this.handleMenuOpen}
							onMenuClose={this.handleMenuClose}
							onView={this.handleView}
							onEdit={this.handleEdit}
							onDescribe={this.handleDescribe}
							onNotification={this.handleNotification}
							onAccess={this.handleAccess}
							onLog={this.handleLog}
							onAddChild={this.handleAddChild}
							onReload={this.handleReload}
							onDelete={this.handleDelete}
							onMove={this.handleMove}
							onValuePopoverOpen={this.handleValuePopoverOpen}
							onValuePopoverClose={this.handleValuePopoverClose}
						/>
					)}
				</Tree>
				{this.state.popover}
				{this.state.dialog}
			</div>
		);
	}

}

export default withStyles(styles)(ConfigTree);
