package tree

type KeyNode struct {
	Value float64
	Key   string
	Left  *KeyNode
	Right *KeyNode
}

type HuffmanTree struct {
	Parent KeyNode
}

func (H *HuffmanTree) SetHuffmanTree(a map[string]float64) {
	(*H).Parent = *(SetHuffmanTree(a, &(H.Parent)))
}

func SetHuffmanTree(a map[string]float64, K *KeyNode) *KeyNode {
	i0, a0 := FindMin(a)
	for {
		if len(a) == 1 {
			break
		}
		delete(a, i0)
		i1, a1 := FindMin(a)
		delete(a, i1)
		K1 := KeyNode{Value: a1, Key: i1}
		K2 := KeyNode{Value: a0 + a1, Key: i0 + "_" + i1}
		K2.Right = K
		K2.Left = &K1
		a[i0+"_"+i1] = a0 + a1
		K = &K2
		i0, a0 = i0+"_"+i1, a0+a1
	}
	return K
}

func FindMin(a map[string]float64) (string, float64) {
	var key string
	var value float64
	k := 0
	for i, j := range a {
		if k == 0 {
			key = i
			value = j
		} else {
			if j < value {
				key = i
				value = j
			}
		}
		k++
	}
	return key, value
}
