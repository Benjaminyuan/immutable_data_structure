package immu_data_structure

import "fmt"

const (
	SHIFT     = 5
	NODE_SIZE = 1 << SHIFT // branching factor,node vector size
	BIT_MASK  = NODE_SIZE - 1
)

type List interface {
	Get(index int) interface{}
	Add(value interface{})
	Remove(index int)
	Size() int
}
type listImpl struct {
	size    uint32
	level   uint32
	root    *trieNode
	tail    *trieNode
	ownerID int64
}

type trieNode struct {
	arr []interface{}
	// last modifyer id
	ownerId int64
}

func updateNode(node *trieNode, level uint32, index int, value interface{}, ownerID int64) *trieNode {
	if value == nil || node == nil || node.arr == nil {
		return node
	}
	idx := (index >> int(level)) & BIT_MASK
	if level > 0 {
		lowerNode := node.arr[idx].(*trieNode)
		// 更新底层节点
		newLowerNode := updateNode(lowerNode, level-SHIFT, index, value, ownerID)
		newNode := editableNode(node, ownerID)
		newNode.arr[idx] = newLowerNode
		return newNode
	}
	newNode := editableNode(node, ownerID)
	newNode.arr[idx] = value
	return newNode
}
func editableNode(node *trieNode, id int64) *trieNode {
	if node.ownerId == id {
		return node
	}
	res := &trieNode{
		ownerId: id,
	}
	if node.arr != nil {
		res.arr = append([]interface{}{}, node.arr...)
	} else {
		res.arr = make([]interface{}, 0)
	}
	return res
}

func (l *listImpl) Get(index int) interface{} {
	if index < 0 || uint32(index) > l.size {
		fmt.Errorf("index [%v] out of boundary", index)
		return nil
	}
	// 查找
	return l.lookUp(uint32(index))
}
func (l *listImpl) Set(index int, value interface{}) {
	// update tail
	if uint32(index) >= l.getTailOffset() {
		l.tail = updateNode(l.tail, 0, index, value, l.ownerID)
	} else {
		l.root = updateNode(l.root, l.level, index, value, l.ownerID)
	}
	// todo,detect whether generate new node
}
func (l *listImpl) Add(value interface{}) {}

func (l *listImpl) Remove(index int) {}
func (l *listImpl) Size() int {
	return 0
}

func NewList() List {
	return &listImpl{}
}

func (l *listImpl) lookUp(index uint32) interface{} {
	node := l.getNode(index)
	if node == nil || node.arr == nil {
		fmt.Errorf("node or node.arr empty for index: %v", index)
		return nil
	}
	return node.arr[index]
}

func (l *listImpl) getNode(index uint32) *trieNode {
	if index > l.getTailOffset() {
		return l.tail
	}
	root := l.root
	for level := l.level; level > 0 && root != nil; level -= SHIFT {
		root = root.arr[(index>>level)&BIT_MASK].(*trieNode)
	}
	return root
}

// getTailOffset: 尾节点优化-> 尾节点存储在 trie tree 的 root
// tailOffset = list.size() - tail.size()
func (l *listImpl) getTailOffset() uint32 {
	// list with only one node
	if l.size < NODE_SIZE {
		return 0
	}
	//  offset = number of node before tail node  * SHIFT
	return ((l.size - 1) >> SHIFT) << SHIFT
}
