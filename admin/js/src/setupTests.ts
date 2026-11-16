// jsdom's Range takes no measurements; CodeMirror, which the text, JSON, YAML
// and template editors mount, measures one while mounting
const emptyRect = { x: 0, y: 0, top: 0, left: 0, bottom: 0, right: 0, width: 0, height: 0, toJSON: () => ({}) } as DOMRect;
Range.prototype.getBoundingClientRect = () => emptyRect;
Range.prototype.getClientRects = () => ({ length: 0, item: () => null, [Symbol.iterator]: [][Symbol.iterator] } as unknown as DOMRectList);

export {};
