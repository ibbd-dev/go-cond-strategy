// 条件策略配置

package crondStrategy

import (
	"sync"
)

// 配置类型
const (
	yesOrNoConf uint8 = iota
	degradeConf
)

// 策略配置结构
type TConf struct {
	// 策略的判断函数
	// check函数返回true时，执行yesAction，否则执行noAction
	// 参数为各统计指标
	Check func(*TEventsMetric) bool

	// 满足条件时需要执行的函数
	YesAction func()

	// 不满足条件是需要执行的函数
	NoAction func()
}

type tConfList struct {
	sync.Mutex
	conf []*TConf
}

var (
	confList *tConfList
)

func init() {
	confList = &tConfList{}
}

// AddConf 增加执行策略配置
func AddConf(conf *TConf) {
	confList.Lock()
	confList.conf = append(confList.conf, conf)
	confList.Unlock()
}
