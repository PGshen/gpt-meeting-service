package datastruct

import (
	"errors"
)

type WithId interface {
	GetId() string
}

type Node[T WithId] struct {
	Data      T        `bson:"data"`
	InDegree  int      `bson:"inDegree"`
	OutDegree int      `bson:"outDegree"`
	Neighbors []string `bson:"neighbors"`
}

func NewNode[T WithId](data T) *Node[T] {
	return &Node[T]{
		Data:      data,
		InDegree:  0,
		OutDegree: 0,
		Neighbors: []string{},
	}
}

func (n *Node[T]) addNeighbor(neighbor T) {
	n.Neighbors = append(n.Neighbors, neighbor.GetId())
}

type Graph[T WithId] struct {
	Nodes []*Node[T] `bson:"nodes"`
}

func NewGraph[T WithId]() *Graph[T] {
	return &Graph[T]{}
}

func (g *Graph[T]) GetNode(data T) *Node[T] {
	for _, n := range g.Nodes {
		if n.Data.GetId() == data.GetId() {
			return n
		}
	}
	return nil
}

func (g *Graph[T]) AddNode(data T) {
	g.Nodes = append(g.Nodes, NewNode(data))
}

// todo 检测是否符合有向无环图
func (g *Graph[T]) AddEdge(from, to T) bool {
	fromNode := g.GetNode(from)
	toNode := g.GetNode(to)
	if fromNode == nil || toNode == nil {
		return false
	}
	fromNode.addNeighbor(to) // 单向
	fromNode.OutDegree++
	toNode.InDegree++
	return true
}

func (g *Graph[T]) RemoveNode(data T) {
	for i, n := range g.Nodes {
		if n.Data.GetId() == data.GetId() { // remove the node from graph's Nodes slice
			g.Nodes = append(g.Nodes[:i], g.Nodes[i+1:]...)
		}
		for j, neighbor := range n.Neighbors {
			if neighbor == data.GetId() { // remove the node from each of its neightbors' neighbor list
				n.Neighbors = append(n.Neighbors[:j], n.Neighbors[j+1:]...)
			}
		}
	}
}

func (g *Graph[T]) RemoveEdge(from, to T) bool {
	fromNode := g.GetNode(from)
	toNode := g.GetNode(to)
	if fromNode == nil || toNode == nil {
		return false
	}
	// Remove toNode from fromNode's neighbor list
	for i, neighbor := range fromNode.Neighbors {
		if neighbor == toNode.Data.GetId() {
			fromNode.Neighbors = append(fromNode.Neighbors[:i], fromNode.Neighbors[i+1:]...)
			fromNode.OutDegree--
			toNode.InDegree--
		}
	}
	return true
}

/**
 * @description: 有向图拓扑排序
 * @return {*}
 */
func (g *Graph[T]) TopologicalSort() ([]T, error) {
	result := make([]T, 0)

	// inegree counter
	counter := make(map[string]int)
	idNodeMap := make(map[string]*Node[T], 0)
	for _, node := range g.Nodes {
		counter[node.Data.GetId()] = 0
		idNodeMap[node.Data.GetId()] = node
	}
	for _, node := range g.Nodes {
		for _, neighbor := range node.Neighbors {
			counter[neighbor]++
		}
	}

	// Enqueue all Nodes with a degree of 0
	queue := make([]string, 0)
	for nodeId, InDegree := range counter {
		if InDegree == 0 {
			queue = append(queue, nodeId)
		}
	}
	for len(queue) > 0 {
		nodeId := queue[0]
		queue = queue[1:]

		node := idNodeMap[nodeId]
		result = append(result, node.Data)

		for _, neighbor := range node.Neighbors {
			counter[neighbor]--
			if counter[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	if len(result) != len(g.Nodes) {
		return nil, errors.New("not DAG")
	}
	return result, nil
}

/**
 * @description: 获取树节点的所有祖先节点
 * @return {*}
 */
func (g *Graph[T]) GetAncestors() map[string][]string {
	nodes := g.Nodes
	// 计算每个节点的父节点
	var nodeParentsMap = make(map[string][]string)
	for _, node := range nodes {
		for _, neighbor := range node.Neighbors {
			if _, ok := nodeParentsMap[neighbor]; !ok {
				nodeParentsMap[neighbor] = []string{}
			}
			nodeParentsMap[neighbor] = append(nodeParentsMap[neighbor], node.Data.GetId())
		}
	}
	// 计算祖先节点
	var nodeAncestorsMap = make(map[string][]string)
	for node := range nodeParentsMap {
		nodeAncestorsMap[node] = getNodeAncestors(node, nodeParentsMap)
	}
	for _, node := range nodes {
		nodeId := node.Data.GetId()
		if _, ok := nodeAncestorsMap[nodeId]; !ok { // 没有祖先节点的，补空
			nodeAncestorsMap[nodeId] = []string{}
		}
	}
	return nodeAncestorsMap
}

/**
 * @description: 获取节点的祖先节点 广度优先
 * @param {string} node
 * @param {map[string][]string} nodeParentsMap
 * @return {*}
 */
func getNodeAncestors(node string, nodeParentsMap map[string][]string) []string {
	queue := []string{node}
	visited := map[string]bool{node: true}
	grandparents := []string{}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, parent := range nodeParentsMap[curr] {
			if !visited[parent] {
				visited[parent] = true
				grandparents = append(grandparents, parent)
				queue = append(queue, parent)
			}
		}
	}
	return grandparents
}
