package interceptors

import (
	"errors"
	"fmt"
)

type tag int

const (
	ToDo tag = iota
	InProgress
	Done
)

type Node struct {
	tag    tag
	entity Interceptor
}

func Sort(entities []Interceptor) ([]Interceptor, error) {
	nodes := make(map[string]*Node, len(entities))
	for _, entity := range entities {
		nodes[entity.Name()] = &Node{tag: ToDo, entity: entity}
	}

	result := make([]Interceptor, 0, len(entities))
	for _, node := range nodes {
		interceptors, err := sortNode(node, nodes)
		if err != nil {
			return nil, err
		}

		result = append(result, interceptors...)
	}

	return result, nil
}

func sortNode(node *Node, nodes map[string]*Node) ([]Interceptor, error) {
	switch node.tag {
	case Done:
		return []Interceptor{}, nil
	case InProgress:
		return nil, errors.New("cycle was found")
	case ToDo:
		result := make([]Interceptor, 0)
		node.tag = InProgress
		for _, name := range node.entity.DependsOn() {
			n, ok := nodes[name]
			if !ok {
				return nil, errors.New("node not found")
			}

			entities, err := sortNode(n, nodes)
			if err != nil {
				return nil, err
			}

			result = append(result, entities...)
		}

		node.tag = Done
		result = append(result, node.entity)

		return result, nil
	default:
		return nil, fmt.Errorf("unknown tag %d", node.tag)
	}
}
