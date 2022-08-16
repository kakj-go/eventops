package dag

import "fmt"

func NewDag() *Dag {
	return &Dag{
		Nodes: map[string]*Node{
			Root: {
				Name: Root,
			},
		},
	}
}

const Root = "root"

type Dag struct {
	Nodes map[string]*Node
}

type Node struct {
	Name      string
	NextNodes []*Node
	NeedNodes []*Node
}

func (dag *Dag) GetAllNeedsTask(name string) (result []string) {
	if name == Root {
		return result
	}
	node := dag.Nodes[name]
	if node == nil {
		return result
	}

	for _, need := range node.NeedNodes {
		result = append(result, need.Name)
		result = append(result, dag.GetAllNeedsTask(need.Name)...)
	}
	return result
}

func (dag *Dag) AddNode(name string) {
	if name == Root {
		return
	}
	dag.Nodes[name] = &Node{
		Name: name,
	}
}

func (dag *Dag) AddEdge(name string, needs []string) error {
	if dag.Nodes[name] == nil {
		return fmt.Errorf("not find node %v", name)
	}
	for _, need := range needs {
		if dag.Nodes[need] == nil {
			return fmt.Errorf("not find need node %v", need)
		}
	}

	node := dag.Nodes[name]
	for _, need := range needs {
		needNode := dag.Nodes[need]
		needNode.NextNodes = append(needNode.NeedNodes, node)
		node.NeedNodes = append(node.NeedNodes, needNode)
	}
	return nil
}

func (dag *Dag) Check() error {
	rootNode := dag.Nodes[Root]
	if rootNode.NeedNodes != nil {
		return fmt.Errorf("the root node cannot depend on other nodes")
	}
	if rootNode.NextNodes == nil {
		return fmt.Errorf("no node depends on root")
	}

	for _, node := range dag.Nodes {
		if node.Name == Root {
			continue
		}
		if node.NeedNodes == nil {
			return fmt.Errorf("node %v Not on the dag dispatch", node.Name)
		}
	}

	var vector = map[string]int{}
	for _, node := range dag.Nodes {
		for _, next := range node.NextNodes {
			_, ok := vector[fmt.Sprintf("%v->%v", node.Name, next.Name)]
			if ok {
				return fmt.Errorf("node %v and next node %v Two nodes will generate a cycle", node.Name, next.Name)
			}
			vector[fmt.Sprintf("%v->%v", node.Name, next.Name)] = 1
		}
	}

	return nil
}
