/*
 * @Descripttion:
 * @version:
 * @Date: 2023-06-13 20:59:40
 * @LastEditTime: 2023-06-19 23:30:33
 */
package datastruct

import "errors"

type TreeNode[T WithId] struct {
	Data     T              `bson:"data"`
	Parent   string         `bson:"parent"`
	Children []*TreeNode[T] `bson:"children"`
}

func NewTreeNode[T WithId](data T, parent string) *TreeNode[T] {
	return &TreeNode[T]{
		Data:     data,
		Parent:   parent,
		Children: []*TreeNode[T]{},
	}
}

type Tree[T WithId] struct {
	Root *TreeNode[T] `bson:"root"`
}

func NewTree[T WithId]() *Tree[T] {
	return &Tree[T]{}
}

func (t *Tree[T]) GetTreeNode(id string, current *TreeNode[T]) *TreeNode[T] {
	if current == nil {
		if t.Root == nil {
			return nil
		}
		// 没有指定current时使用根节点
		current = t.Root
	}
	if current.Data.GetId() == id {
		return current
	}

	for _, child := range current.Children {
		if treeNode := t.GetTreeNode(id, child); treeNode != nil {
			return treeNode
		}
	}
	return nil
}

func (t *Tree[T]) Insert(parentId string, data T) error {
	if t.Root == nil {
		if parentId != "" {
			return errors.New("can't insert node on non-existent parent")
		}
		t.Root = NewTreeNode(data, "")
		return nil
	}
	parentNode := t.GetTreeNode(parentId, t.Root)
	if parentNode == nil {
		return errors.New("can't insert node on non-existent parent")
	}
	parentNode.Children = append(parentNode.Children, NewTreeNode(data, parentId))
	return nil
}

func (t *Tree[T]) Remove(id string) error {
	if t.Root == nil {
		return errors.New("can't remove node from empty tree")
	}
	treeNode := t.GetTreeNode(id, t.Root)
	if treeNode == nil {
		return errors.New("can't remove non-existent node")
	}
	if treeNode == t.Root {
		t.Root = nil
		return nil
	}
	parent := treeNode.Parent
	parentNode := t.GetTreeNode(parent, nil)
	children := parentNode.Children
	for i, child := range children {
		if child == treeNode {
			copy(children[i:], children[i+1:])
			parentNode.Children = children[:len(children)-1]
			treeNode.Parent = ""
			treeNode.Children = nil
			return nil
		}
	}
	return nil
}

/**
 * @description: 更新节点（ID是不可变的）
 * @param {T} data
 * @return {*}
 */
func (t *Tree[T]) Update(id string, data T) error {
	target := t.GetTreeNode(id, t.Root)
	if target == nil {
		return errors.New("can't update non-existen node")
	}
	target.Data = data
	return nil
}

/**
 * @description: 获取从根节点到指定节点的路径
 * @param {*TreeNode[T]} node
 * @param {string} id
 * @return {*}
 */
func (t *Tree[T]) FindPath(node *TreeNode[T], id string) (path []*TreeNode[T], found bool) {
	if node == nil {
		node = t.Root
	}
	if node == nil {
		return []*TreeNode[T]{}, false
	}
	if node.Data.GetId() == id {
		return []*TreeNode[T]{node}, true
	} else if len(node.Children) > 0 {
		for _, child := range node.Children {
			if childPath, ok := t.FindPath(child, id); ok {
				return append([]*TreeNode[T]{node}, childPath...), true
			}
		}
		return []*TreeNode[T]{}, false
	}
	return []*TreeNode[T]{}, false
}

/**
 * @description: 合并树
 * @param {*Tree[T]} t1
 * @return {*}
 */
func (t *Tree[T]) MergeTree(t1 *Tree[T]) *Tree[T] {
	newRoot := t.mergeNodes(t.Root, t1.Root)
	return &Tree[T]{newRoot}
}

func (t *Tree[T]) mergeNodes(node1, node2 *TreeNode[T]) *TreeNode[T] {
	if node1 == nil && node2 == nil {
		return nil // 节点都为 nil，返回空
	}

	// 创建新节点
	mergedNode := &TreeNode[T]{}
	if node1 != nil {
		mergedNode.Data = node1.Data
		mergedNode.Parent = node1.Parent
	} else {
		mergedNode.Data = node2.Data
		mergedNode.Parent = node2.Parent
	}

	// 合并子节点
	children := make(map[string]*TreeNode[T])
	if node1 != nil {
		for _, child := range node1.Children {
			children[child.Data.GetId()] = child
		}
	}
	if node2 != nil {
		for _, child := range node2.Children {
			if _, exist := children[child.Data.GetId()]; exist { // ID相同时，node1覆盖node2
				mergedChild := t.mergeNodes(children[child.Data.GetId()], child)
				children[child.Data.GetId()] = mergedChild
			} else {
				children[child.Data.GetId()] = child
			}
		}
	}
	for _, child := range children {
		mergedNode.Children = append(mergedNode.Children, child)
	}
	return mergedNode
}
