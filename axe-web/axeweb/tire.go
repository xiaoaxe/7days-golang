//trie
//@author: baoqiang
//@time: 2021/10/15 21:49:04
package axeweb

import (
	"fmt"
	"strings"
)

type node struct {
	pattern  string
	part     string
	children []*node
	isWild   bool // is wildcard model
}

func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

func (n *node) insert(pattern string, parts []string, height int) {
	// only leaf node has pattern val
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	// not found
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		// add one
		n.children = append(n.children, child)
	}
	// interater next level
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	// found end
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _, c := range children {
		result := c.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

// get all paths
func (n *node) travel(list *([]*node)) {
	if n.pattern != "" {
		*list = append(*list, n)
	}

	for _, c := range n.children {
		c.travel(list)
	}
}

func (n *node) matchChild(part string) *node {
	for _, c := range n.children {
		if c.part == part || c.isWild {
			return c
		}
	}
	return nil
}

func (n *node) matchChildren(part string) []*node {
	var nodes = make([]*node, 0)
	for _, c := range n.children {
		if c.part == part || c.isWild {
			nodes = append(nodes, c)
		}
	}
	return nodes
}
