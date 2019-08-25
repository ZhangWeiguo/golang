package abtest

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/spaolacci/murmur3"
	"io/ioutil"
	"math"
	"os"
)

const (
	MinBucketNum = 1000
)

type abTestConfig struct {
	XMLName   xml.Name `xml:"experiment"`
	BucketNum uint64   `xml:"bucket_num"`
	Exps      []exp    `xml:"exp"`
}

type exp struct {
	XMLName xml.Name `xml:"exp"`
	ExpId   int      `xml:"exp_id,attr"`
	ExpName string   `xml:"exp_name,attr"`
	ExpRate float64  `xml:"exp_rate,attr"`
	Layered bool     `xml:"layered,attr"`
	Layers  []layer  `xml:"layer"`
}

type layer struct {
	XMLName   xml.Name   `xml:"layer"`
	LayerId   int        `xml:"layer_id,attr"`
	LayerName string     `xml:"layer_name,attr"`
	SubLayers []sublayer `xml:"layer"`
}

type sublayer struct {
	XMLName   xml.Name `xml:"layer"`
	LayerId   int      `xml:"layer_id,attr"`
	LayerName string   `xml:"layer_name,attr"`
	LayerRate float64  `xml:"layer_rate,attr"`
}

type ExpRange struct {
	id      int
	name    string
	layered bool
	val     uint64
}

type LayerRange struct {
	id   int
	name string
	val  uint64
}

type LayerDict struct {
	id    int
	name  string
	rates []LayerRange
}

type ABTest struct {
	config         abTestConfig
	bucketNum      uint64
	Logger         func(s string)
	expRange       []ExpRange
	layerRangeDict map[int][]LayerDict
}

func (ab *ABTest) Init(path string) (e error) {
	if ab.Logger == nil {
		ab.Logger = func(s string) {
			fmt.Println(s)
		}
	}
	e = ab.ParseXML(path)
	return e
}

func (ab *ABTest) ParseXML(path string) (e error) {
	reader, e := os.OpenFile(path, os.O_RDONLY, 0)
	if e == nil {
		data, e := ioutil.ReadAll(reader)
		if e == nil {
			e = xml.Unmarshal(data, &ab.config)
			if e == nil {
				e = ab.check()
				if e == nil {
					ab.transform()
				}
			}
		}
	}
	return e
}

func (ab *ABTest) Print() {
	ab.Logger(fmt.Sprintf("ABTestCconfig(%s) BucketNUm(%d) Start", ab.config.XMLName, ab.config.BucketNum))
	ab.Logger("======================================================")
	for _, exp := range ab.config.Exps {
		ab.Logger(fmt.Sprintf("ExpName(%s) ExpId(%d) ExpRate(%3.3f) Layered(%v)",
			exp.ExpName, exp.ExpId, exp.ExpRate, exp.Layered))
		if exp.Layered {
			for _, layer := range exp.Layers {
				for _, sublayer := range layer.SubLayers {
					ab.Logger(fmt.Sprintf(">> LayerName(%s) LayerId(%d) SubLayerName(%s) SubLayerID(%d) SubLayerRate(%3.3f)",
						layer.LayerName, layer.LayerId, sublayer.LayerName, sublayer.LayerId, sublayer.LayerRate))
				}
			}
		}
		ab.Logger("------------------------------------------------------")
	}
}

func (ab *ABTest) check() (e error) {
	if ab.config.BucketNum < MinBucketNum {
		return errors.New("To Small BucketNum ")
	} else {
		ab.bucketNum = ab.config.BucketNum
		var expSum, layerSum float64
		expSum = 0.0
		for _, exp := range ab.config.Exps {
			expSum += exp.ExpRate
			if exp.Layered {
				for _, layer := range exp.Layers {
					layerSum = 0.0
					for _, sublayer := range layer.SubLayers {
						layerSum += sublayer.LayerRate
					}
					if math.Abs(layerSum-1.0) >= 1e-10 {
						return errors.New(fmt.Sprintf("Layer(%s) Sum Not Equals 100 ", layer.LayerName))
					}
				}
			}
		}
		if math.Abs(expSum-1.0) >= 1e-10 {
			return errors.New("Exp Sum Not Equals 100% ")
		}
	}
	return nil
}

func (ab *ABTest) transform() {
	ab.expRange = make([]ExpRange, 0)
	ab.layerRangeDict = make(map[int][]LayerDict)
	var expSum, layerSum uint64
	expSum = 0.0
	for _, exp := range ab.config.Exps {
		expSum += uint64(exp.ExpRate * float64(ab.bucketNum))
		ab.expRange = append(ab.expRange, ExpRange{
			exp.ExpId, exp.ExpName, exp.Layered, expSum})
		if exp.Layered {
			ab.layerRangeDict[exp.ExpId] = make([]LayerDict, 0)
			for _, layer := range exp.Layers {
				sublayerRange := make([]LayerRange, 0)
				layerSum = 0.0
				for _, sublayer := range layer.SubLayers {
					layerSum += uint64(sublayer.LayerRate * float64(ab.bucketNum))
					sublayerRange = append(sublayerRange,
						LayerRange{sublayer.LayerId, sublayer.LayerName, layerSum})
				}
				ab.layerRangeDict[exp.ExpId] = append(ab.layerRangeDict[exp.ExpId], LayerDict{
					layer.LayerId, layer.LayerName, sublayerRange})
			}
		}
	}
}

func (ab *ABTest) GetTag(id string) (int, string, bool, []string) {
	expId, expName, layered, bucket := ab.getExp(id)
	layers := make([]string, 0)
	if layered {
		layers = ab.getLayers(expId, bucket)
	}
	return expId, expName, layered, layers
}

func (ab *ABTest) getExp(id string) (int, string, bool, uint64) {
	bucket := ab.hash(id) % ab.bucketNum
	for _, exp := range ab.expRange {
		if bucket <= exp.val {
			return exp.id, exp.name, exp.layered, bucket
		}
	}
	return 0, "", false, 0
}

func (ab *ABTest) getLayers(expId int, bucket uint64) (s []string) {
	s = make([]string, 0)
	for _, val := range ab.layerRangeDict[expId] {
		newBucket := ab.hash(fmt.Sprintf("%d%d", val.id, bucket)) % ab.bucketNum
		for _, layer := range val.rates {
			if newBucket <= layer.val {
				s = append(s, fmt.Sprintf("%d,%s:%d,%s", val.id, val.name, layer.id, layer.name))
				break
			}
		}
	}
	return s
}

func (ab *ABTest) hash(id string) uint64 {
	h64Byte := murmur3.New64()
	_, e := h64Byte.Write([]byte(id))
	if e != nil {
		ab.Logger(fmt.Sprintf("Hash Id(%s) Error", id))
	}
	hash := h64Byte.Sum64()
	return hash
}
