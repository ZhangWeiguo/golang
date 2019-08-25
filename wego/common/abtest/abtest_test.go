package abtest

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type result struct {
	Id   int
	Name string
	Val  int
}

func BenchmarkABTest(b *testing.B) {
	b.StopTimer()
	ab := ABTest{}
	e := ab.Init("abtest.xml")
	if e == nil {
		id := strconv.FormatUint(rand.Uint64(), 36) + strconv.FormatInt(int64(10), 8)
		b.StartTimer()
		_, _, _, _ = ab.GetTag(id)
	}
}

func TestABTest(t *testing.T) {
	var ab ABTest
	var N int
	var expData map[int]*result
	expData = make(map[int]*result, 0)
	layerData := make(map[int]map[string]int)
	N = 1000000
	e := ab.Init("abtest.xml")
	if e == nil {
		ab.Print()
		ids := make([]string, 0)
		for i := 0; i < N; i++ {
			ids = append(ids, strconv.FormatUint(rand.Uint64(), 36)+strconv.FormatInt(int64(i), 8))
		}
		t0 := time.Now().Nanosecond()
		for _, id := range ids {
			expId, expName, layered, layers := ab.GetTag(id)
			if _, ok := expData[expId]; ok {
				expData[expId].Val += 1
			} else {
				expData[expId] = &result{expId, expName, 1}
			}
			if layered {
				if _, ok := layerData[expId]; !ok {
					layerData[expId] = make(map[string]int)
				}
				for _, layer := range layers {
					if _, ok := layerData[expId][layer]; ok {
						layerData[expId][layer] += 1
					} else {
						layerData[expId][layer] = 1
					}
				}
			}
		}
		t1 := time.Now().Nanosecond()
		fmt.Println("ABTest Result Validate")
		fmt.Println("======================================================")
		for _, val := range expData {
			fmt.Println(fmt.Sprintf("ExpName(%s) ExpId(%d) ExpRate(%3.3f)",
				val.Name, val.Id, float64(val.Val)/float64(N)))
			if r, ok := layerData[val.Id]; ok {
				for layerinfo, val := range r {
					fmt.Println(fmt.Sprintf(">>[%25s] %3.3f", layerinfo, float64(val)/float64(N)))
				}
			}
			fmt.Println("------------------------------------------------------")
		}
		fmt.Println(fmt.Sprintf("All Cost (%3.4fms) Avg Cost (%3.4fns) (%d / S)",
			float64(t1-t0)/1000000.0, float64(t1-t0)/float64(N), int64(float64(N)/(float64(t1-t0)/1000000000))))
	} else {
		fmt.Println("ABTest Init Error")
	}
}
