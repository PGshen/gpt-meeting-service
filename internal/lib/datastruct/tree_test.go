/*
 * @Date: 2023-06-14 00:09:54
 * @LastEditors: Please set LastEditors
 * @LastEditTime: 2023-06-14 13:44:21
 * @FilePath: /gpt-meeting-service/internal/lib/datastruct/tree_test.go
 */
package datastruct

import (
	"fmt"
	"testing"
)

type treeData struct {
	Id   string
	Data string
}

func (s treeData) GetId() string {
	return s.Id
}

func NewTd(id, data string) treeData {
	return treeData{
		Id:   id,
		Data: data,
	}
}

func TestTree(t *testing.T) {
	tree := NewTree[treeData]()

	a := NewTd("1", "A")
	b := NewTd("2", "B")
	c := NewTd("3", "C")
	d := NewTd("4", "D")
	e := NewTd("5", "E")
	f := NewTd("6", "F")
	tree.Insert("", a)
	tree.Insert("1", b)
	tree.Insert("1", c)
	tree.Insert("1", d)
	tree.Insert("4", e)
	tree.Insert("4", f)

	fmt.Println(tree.GetTreeNode("4", nil))
	err := tree.Remove("D")
	fmt.Println(err)
	fmt.Println(tree.GetTreeNode("E", nil))
	tree.Update("2", NewTd("22", "BB"))
	fmt.Printf("tree: %v", tree)
	path, _ := tree.FindPath(nil, "5")
	fmt.Printf("path: %v", path)
}
