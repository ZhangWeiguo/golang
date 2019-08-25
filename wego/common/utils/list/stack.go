package list

type Stack struct {
	data []int
	Size int
}

func (d *Stack) Init() {
	d.data = make([]int, 0, 100)
	d.Size = 0
}

func (d *Stack) InitArray(a []int) {
	d.Init()
	copy(d.data, a)
	d.Size = len(a)

}

func (d *Stack) Push(a int) {
	d.Size++
	d.data = append(d.data, a)
}

func (d *Stack) Pop() {
	N := d.Size
	d.data = d.data[0 : N-1]
	d.Size--
}
