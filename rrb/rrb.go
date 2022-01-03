package rrb

const NODE_BITS = 5
const NODE_SIZE = 1 << NODE_BITS
const INDEX_MARSK = NODE_SIZE - 1

type RRBVector interface {
	Apply(int)
	Updated()
	Append()
	Prepend()
	Drop()
	Take()
}
type Node struct {
	ele  []interface{}
	l    int
	mark int32
}

func NewNode(nodes ...interface{}) *Node {
	ele := make([]interface{}, NODE_SIZE)
	copy(ele, nodes)
	return &Node{
		ele:  ele,
		l:    len(nodes) - 1,
		mark: ((1 << len(nodes)) - 1) << (NODE_SIZE - len(nodes)),
	}
}

func (node *Node) GetNode(idx int) *Node {
	if val, ok := node.ele[idx].(*Node); ok {
		return val
	}
	return nil
}
func (node *Node) Get(idx int) interface{} {
	return node.ele[idx]
}
func (node *Node) Set(idx int, val interface{}) {
	node.ele[idx] = val
	node.mark |= 1 << (NODE_SIZE - 1 - idx)
}
func (node *Node) Clone() *Node {
	newNode := &Node{ele: make([]interface{}, NODE_SIZE), l: node.l, mark: node.mark}
	copy(newNode.ele, node.ele)
	return newNode
}
func (node *Node) Add(ele interface{}) {
	node.l += 1
	node.Set(node.l, ele)
}

func (node *Node) Last() *Node {
	for i := 0; i < NODE_SIZE; i++ {
		if node.mark&(1<<i) == 1 {
			if val, ok := node.ele[i].(*Node); ok {
				return val
			}
		}
	}
	return nil
}

type RRbVectorImpl struct {
	root     *Node
	depth    int
	startIdx int
	endIdx   int
}

// Apply: applyoperation returns the element located at the given inde
func (vec *RRbVectorImpl) Apply(index int) interface{} {
	var getElem func(node *Node, depth int) interface{}
	getElem = func(node *Node, depth int) interface{} {
		idxInNode := getNodeIdx(index, depth, vec.startIdx)
		if depth == 1 {
			return node.Get(idxInNode)
		}
		return getElem(node.GetNode((idxInNode)), depth-1)
	}
	return getElem(vec.root, vec.depth)
}

func getNodeIdx(index int, depth int, startIdx int) int {
	return ((index + startIdx) >> (NODE_BITS * (depth - 1))) & INDEX_MARSK
}

func (vec *RRbVectorImpl) Updated(index int, ele interface{}) *RRbVectorImpl {
	var updateNode func(node *Node, depth int) *Node
	updateNode = func(node *Node, depth int) *Node {
		idxInNode := getNodeIdx(index, depth, vec.startIdx)
		// 拷贝
		newNode := node.Clone()
		if depth == 1 {
			newNode.Set(idxInNode, ele)
		} else {
			newNode.Set(idxInNode, updateNode(node.GetNode(idxInNode), depth-1))
		}
		return newNode
	}
	return &RRbVectorImpl{
		root:  updateNode(vec.root, vec.depth),
		depth: vec.depth,
	}
}
func copyLeafAddAppend(node *Node, ele interface{}) *Node {
	newNode := node.Clone()
	newNode.Add(ele)
	return newNode
}
func (vec *RRbVectorImpl) Append(ele interface{}) *RRbVectorImpl {
	var newBranch func(depth int) *Node
	newBranch = func(depth int) *Node {
		if depth == 1 {
			return NewNode(ele)
		}
		return NewNode(newBranch(depth - 1))
	}
	var isTreeFull func(node *Node, depth int) bool
	isTreeFull = func(node *Node, depth int) bool {
		if (1<<(depth*NODE_BITS)-1)&vec.endIdx == 1 {
			return true
		}
		return false
	}
	var copyAddUpdateLast func(node *Node, last *Node) *Node
	copyAddUpdateLast = func(node *Node, last *Node) *Node {
		newNode := node.Clone()
		newNode.Set(newNode.size-1, last)
		return newNode
	}
	var copyBranchAndAppend func(node, branch *Node) *Node
	copyBranchAndAppend = func(node, branch *Node) *Node {
		newNode := node.Clone()
		newNode.Add(branch)
		return newNode
	}
	var appended func(node *Node, depth int) *Node
	appended = func(node *Node, depth int) *Node {
		if depth == 1 {
			return copyLeafAddAppend(node, ele)
			// last node not full
		} else if !isTreeFull(node.Last(), depth-1) {
			return copyAddUpdateLast(node, appended(node.Last(), depth-1))
		} else {
			// last node full
			return copyBranchAndAppend(node, newBranch(depth-1))
		}
	}
	// root node full check
	if !isTreeFull(vec.root, vec.depth) {
		return &RRbVectorImpl{
			root:     appended(vec.root, vec.depth),
			depth:    vec.depth,
			startIdx: vec.startIdx,
			endIdx:   vec.endIdx + 1,
		}
	}
	// root node full generate no
	return &RRbVectorImpl{
		root:     NewNode(vec.root, newBranch(vec.depth)),
		depth:    vec.depth + 1,
		startIdx: vec.startIdx,
		endIdx:   vec.endIdx + 1,
	}

}
func (vec *RRbVectorImpl) Prepend(ele interface{}) *RRbVectorImpl {
	copyAndUpdate := func(node *Node, idxInNode int, ele interface{}) *Node {
		newNode := node.Clone()
		newNode.Set(idxInNode, ele)
		return newNode
	}
	var prepend func(node *Node, depth int) *Node
	prepend = func(node *Node, depth int) *Node {
		idxInNode := getNodeIdx(vec.startIdx-1, depth, 0)
		if depth == 1 {
			return copyAndUpdate(node, idxInNode, ele)
		}
		return copyAndUpdate(node, idxInNode, prepend(node.GetNode(idxInNode), depth-1))
	}
	var newBranch func(depth int) *Node
	newBranch = func(depth int) *Node {
		newNode := NewNode()
		if depth == 1 {
			newNode.Set(INDEX_MARSK, ele)
		} else {
			newNode.Set(INDEX_MARSK, newBranch(depth-1))
		}
		return newNode
	}
	// start with zero
	// case: root is full
	// case-2:
	if vec.startIdx == 0 {
		return &RRbVectorImpl{
			root:     NewNode(newBranch(vec.depth), vec.root),
			depth:    vec.depth + 1,
			startIdx: (1 << (NODE_BITS * vec.depth)) - 1,
			endIdx:   (1 << (NODE_BITS * vec.depth)) + vec.endIdx,
		}
	}
	// start with
	return &RRbVectorImpl{
		depth:    vec.depth,
		root:     prepend(vec.root, vec.depth),
		startIdx: vec.startIdx - 1,
		endIdx:   vec.endIdx,
	}
}
