package utils

import (
    "crypto/sha256"
    "encoding/hex"
)

type Node struct {
    Left  *Node
    Right *Node
    Data  string
}

type MerkleTree struct {
    Leaves []*Node
    Levels [][]*Node
    Root   *Node
}

func NewMerkleTree(dataBlocks []string) *MerkleTree {
    var leaves []*Node
    for _, data := range dataBlocks {
        leaves = append(leaves, &Node{Data: hash(data)})
    }
    
    tree := &MerkleTree{}
    tree.Root = tree.buildTree(leaves)
    tree.Leaves = leaves
    return tree
}

func hash(data string) string {
    h := sha256.New()
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

func getParent(node1, node2 *Node) *Node{
    if node2 == nil {
        node2 = node1
    }
    concat := node1.Data + node2.Data
    parent := &Node{
        Data: hash(concat), 
        Left: node1, 
        Right: node2,
    }
    return parent
}

func (t *MerkleTree) buildTree(nodes []*Node) *Node{
    if len(nodes) < 1{
        return nil
    }

    t.Levels = append(t.Levels, nodes)

    for len(nodes) > 1{
        var nextLevel []*Node
        for i := 0; i < len(nodes); i+=2 {
            var parent *Node
            if i+1 < len(nodes){
                parent = getParent(nodes[i], nodes[i+1])
            } else {
                parent = getParent(nodes[i], nil)
            }
            nextLevel = append(nextLevel, parent)
        }

        nodes = nextLevel
        t.Levels = append(t.Levels, nodes)
    }

    return nodes[0]
}

func (t *MerkleTree) GetRootHash() string {
    if t.Root == nil {
        return ""
    }
    return t.Root.Data
}