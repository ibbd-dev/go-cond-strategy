// 降维配置

package servicesDegrade

import (
	"sync"
)

// check函数返回true时，执行yesAction，否则执行noAction
type IConf interface {
	// 判断是否满足条件的函数
	Check(*TEventsMetric) bool

	// 满足条件时需要执行的函数
	YesAction()

	// 不满足条件是需要执行的函数
	NoAction()
}

type TConf struct {
	IConf
}

type TConfList struct {
	sync.Mutex
	Conf []*TConf
}

var (
	confList *TConfList
)

func init() {
	//confList = make([]*TConf)
}

// AddConf 增加服务降维配置
func AddConf(conf *TConf) {
	confList.Lock()
	confList.Conf = append(confList.Conf, conf)
	confList.Unlock()
}
