// 服务降级策略

package crondStrategy

import (
	"sync"
)

// 服务降级的类型
type DegradeLevel uint8

// 5个级别的类型
const (
	DegradeL0 DegradeLevel = iota
	DegradeL1
	DegradeL2
	DegradeL3
	DegradeL4
)

const LevelCount = 5

// 策略配置结构
type TDegradeConf struct {
	// 最后的level
	lastLevel DegradeLevel

	// 降级策略的判断函数
	// 参数为各统计指标
	Check func(*TEventsMetric) DegradeLevel

	// 判断之后执行的函数
	// 每个级别对应一个函数
	actions [LevelCount]func()
}

type tDegradeConfList struct {
	sync.Mutex
	conf []*TDegradeConf
}

var (
	degConfList *tDegradeConfList
)

func init() {
	degConfList = &tDegradeConfList{}
}

// AddConf 增加执行策略配置
func AddDegradeConf(conf *TDegradeConf) {
	degConfList.Lock()
	degConfList.conf = append(degConfList.conf, conf)
	degConfList.Unlock()
}

// 增加级别的执行函数
func (d *TDegradeConf) AddDegradeAction(level DegradeLevel, action func()) {
	d.actions[level] = action
}
